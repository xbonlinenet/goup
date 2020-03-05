package cache

import (
	"errors"
	"log"
	"sync/atomic"
	"time"

	"github.com/xbonlinenet/goup/frame/cache/internal/lrucache"
	"github.com/xbonlinenet/goup/frame/cache/internal/singleflight"
	"github.com/xbonlinenet/goup/frame/util"
	"fmt"
	"github.com/go-redis/redis"
	"reflect"
)

var ErrCacheMiss = errors.New("cache: key is missing")
var errRedisLocalCacheNil = errors.New("cache: both Redis and LocalCache are nil")

type rediser interface {
	Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(key string) *redis.StringCmd
	Del(keys ...string) *redis.IntCmd
}

type Item struct {
	Key    string
	Object interface{}

	// Func returns object to be cached.
	Func func() (interface{}, error)

	// Expiration is the cache expiration time.
	// Default expiration is 1 hour.
	Expiration time.Duration
}

func (item *Item) object() (interface{}, error) {
	if item.Object != nil {
		return item.Object, nil
	}
	if item.Func != nil {
		return item.Func()
	}
	return nil, nil
}

func (item *Item) exp() time.Duration {
	if item.Expiration < 0 {
		return 0
	}
	if item.Expiration < time.Second {
		return time.Hour
	}
	return item.Expiration
}

type Codec struct {
	Redis rediser

	localCache *lrucache.Cache

	Marshal   func(interface{}) ([]byte, error)
	Unmarshal func([]byte, interface{}) error

	group singleflight.Group

	hits        uint64
	misses      uint64
	localHits   uint64
	localMisses uint64
}

// UseLocalCache causes Codec to cache items in local LRU cache.
func (cd *Codec) UseLocalCache(maxLen int, expiration time.Duration) {
	cd.localCache = lrucache.New(maxLen, expiration)
}

// Set caches the item.
func (cd *Codec) Set(item *Item) error {
	_, err := cd.set(item)
	return err
}

func (cd *Codec) set(item *Item) ([]byte, error) {
	object, err := item.object()
	if err != nil {
		return nil, err
	}

	b, err := cd.Marshal(object)
	if err != nil {
		log.Printf("cache: Marshal key=%q failed: %s", item.Key, err)
		return nil, err
	}

	if cd.localCache != nil {
		cd.localCache.Set(item.Key, b)
	}

	if cd.Redis == nil {
		if cd.localCache == nil {
			return nil, errRedisLocalCacheNil
		}
		return b, nil
	}

	err = cd.Redis.Set(item.Key, b, item.exp()).Err()
	if err != nil {
		log.Printf("cache: Set key=%q failed: %s", item.Key, err)
		return nil, err
	}
	return b, nil
}

// Exists reports whether object for the given key exists.
func (cd *Codec) Exists(key string) bool {
	return cd.Get(key, nil) == nil
}

// Get gets the object for the given key.
func (cd *Codec) Get(key string, object interface{}) error {
	return cd.get(key, object, false)
}

func (cd *Codec) get(key string, object interface{}, onlyLocalCache bool) error {
	b, err := cd.getBytes(key, onlyLocalCache)
	if err != nil {
		return err
	}

	if object == nil || len(b) == 0 {
		return nil
	}

	err = cd.Unmarshal(b, object)
	if err != nil {
		log.Printf("cache: key=%q Unmarshal(%T) failed: %s", key, object, err)
		return err
	}

	return nil
}

func (cd *Codec) getBytes(key string, onlyLocalCache bool) ([]byte, error) {
	if cd.localCache != nil {
		b, ok := cd.localCache.Get(key)
		if ok {
			atomic.AddUint64(&cd.localHits, 1)
			return b, nil
		}
		atomic.AddUint64(&cd.localMisses, 1)
	}

	if onlyLocalCache {
		return nil, ErrCacheMiss
	}
	if cd.Redis == nil {
		if cd.localCache == nil {
			return nil, errRedisLocalCacheNil
		}
		return nil, ErrCacheMiss
	}

	b, err := cd.Redis.Get(key).Bytes()
	if err != nil {
		atomic.AddUint64(&cd.misses, 1)
		if err == redis.Nil {
			return nil, ErrCacheMiss
		}
		log.Printf("cache: Get key=%q failed: %s", key, err)
		return nil, err
	}
	atomic.AddUint64(&cd.hits, 1)

	if cd.localCache != nil {
		cd.localCache.Set(key, b)
	}
	return b, nil
}

// Once gets the item.Object for the given item.Key from the cache or
// executes, caches, and returns the results of the given item.Func,
// making sure that only one execution is in-flight for a given item.Key
// at a time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results.
func (cd *Codec) Once(item *Item) error {
	b, cached, err := cd.getItemBytesOnce(item)
	if err != nil {
		return err
	}

	if item.Object == nil || len(b) == 0 {
		return nil
	}

	err = cd.Unmarshal(b, item.Object)
	if err != nil {
		log.Printf("cache: key=%q Unmarshal(%T) failed: %s", item.Key, item.Object, err)
		if cached {
			_ = cd.Delete(item.Key)
			return cd.Once(item)
		} else {
			return err
		}
	}

	return nil
}

func (cd *Codec) getItemBytesOnce(item *Item) (b []byte, cached bool, err error) {
	if cd.localCache != nil {
		b, err := cd.getItemBytesFast(item)
		if err == nil {
			return b, true, nil
		}
	}

	obj, err := cd.group.Do(item.Key, func() (interface{}, error) {
		b, err := cd.getItemBytes(item)
		if err == nil {
			cached = true
			return b, nil
		}

		obj, err := item.Func()
		if err != nil {
			return nil, err
		}

		b, err = cd.set(&Item{
			Key:        item.Key,
			Object:     obj,
			Expiration: item.Expiration,
		})
		return b, err
	})
	if err != nil {
		return nil, false, err
	}
	return obj.([]byte), cached, nil
}

func (cd *Codec) getItemBytes(item *Item) ([]byte, error) {
	return cd.getBytes(item.Key, false)
}

func (cd *Codec) getItemBytesFast(item *Item) ([]byte, error) {
	return cd.getBytes(item.Key, true)
}

func (cd *Codec) Delete(key string) error {
	if cd.localCache != nil {
		cd.localCache.Delete(key)
	}

	if cd.Redis == nil {
		if cd.localCache == nil {
			return errRedisLocalCacheNil
		}
		return nil
	}

	deleted, err := cd.Redis.Del(key).Result()
	if err != nil {
		log.Printf("cache: Del key=%q failed: %s", key, err)
		return err
	}
	if deleted == 0 {
		return ErrCacheMiss
	}
	return nil
}

type Stats struct {
	Hits        uint64
	Misses      uint64
	LocalHits   uint64
	LocalMisses uint64
}

// Stats returns cache statistics.
func (cd *Codec) Stats() *Stats {
	stats := Stats{
		Hits:   atomic.LoadUint64(&cd.hits),
		Misses: atomic.LoadUint64(&cd.misses),
	}
	if cd.localCache != nil {
		stats.LocalHits = atomic.LoadUint64(&cd.localHits)
		stats.LocalMisses = atomic.LoadUint64(&cd.localMisses)
	}
	return &stats
}

//
// 基类，redis缓存和进程内缓存都是一致的写法
type TertiaryCache struct {
	CacheKeyFormatFunc func(interface{}) string
	// lru
	LRUExpire, RedisCacheExpire int64

	// lru
	localCache *lrucache.Cache

	// redis
	redisCmd redis.Cmdable

	// 序列化
	Marshal   func(interface{}) ([]byte, error)
	Unmarshal func([]byte, interface{}) error
}

func convertList(idList []interface{}, idResM map[interface{}]interface{}) []interface{} {
	res := make([]interface{}, 0, len(idList))
	for _, id := range idList {
		if v, ok := idResM[id]; ok {
			res = append(res, v)
		} else {
			res = append(res, nil)
		}

	}

	return res
}

//
//// 存放结果
//type TertiaryCacheItem interface {
//}

/*
makeRspFunc 自己提供构建map函数
rsp map[interface{}]interface{} 必须是已经存好数据的map,只允许修改value指针的值
*/
func exampleTertiaryCache() {
	tc := GetTertiaryCache("normal")
	type Item struct {
		Name string
		Age  string
	}

	tc.List(&TertiaryItem{
		func(id interface{}) string {
			return fmt.Sprint("test:%d", id)
		},
		time.Second * 12,
		func(idList []interface{}, rsp map[interface{}]interface{}) {
			for _, id := range idList {
				rsp[id] = &Item{"a", "b"}
			}
		},
		[]interface{}{12, 2, 3},
		reflect.TypeOf(1),
	})

}

type TertiaryItem struct {
	CacheKeyFormatFunc func(interface{}) string
	RedisExpiration    time.Duration
	GetRemoteData      func(idList []interface{}, rsp map[interface{}]interface{})
	IdList             []interface{}
	DataType           reflect.Type
}

func (b TertiaryCache) List(item *TertiaryItem) []interface{} {

	b.CacheKeyFormatFunc = item.CacheKeyFormatFunc
	idList := item.IdList
	idResM := make(map[interface{}]interface{}, len(idList))
	// 1.进程内缓存
	failedList1 := b.ProcessCache(idList, idResM)
	//fmt.Println("ProcessCache failedList1:", failedList1)
	if len(failedList1) == 0 {
		return convertList(idList, idResM)
	}

	// 2.redis 缓存
	failedList2, success2 := b.RedisCache(failedList1, idResM, item.DataType)
	//fmt.Println("RedisCache failedList2:", failedList2)
	b.saveProcessCache(success2, idResM)
	if len(failedList2) == 0 {
		return convertList(idList, idResM)
	}

	if item.GetRemoteData != nil {
		// 3.查询mysql,mongodb,grpc,http api等方式
		item.GetRemoteData(failedList2, idResM)
		//fmt.Println("GetRemoteData idResM:%v", idResM)
		b.saveRedisCache(failedList2, idResM, item.RedisExpiration)
	}

	return convertList(idList, idResM)
}

func (b TertiaryCache) Map(item *TertiaryItem) map[interface{}]interface{} {

	b.CacheKeyFormatFunc = item.CacheKeyFormatFunc
	idList := item.IdList
	idResM := make(map[interface{}]interface{}, len(idList))
	// 1.进程内缓存
	failedList1 := b.ProcessCache(idList, idResM)
	if len(failedList1) == 0 {
		return idResM
	}

	// 2.redis 缓存
	failedList2, success2 := b.RedisCache(failedList1, idResM, item.DataType)
	b.saveProcessCache(success2, idResM)
	if len(failedList2) == 0 {
		return idResM
	}

	if item.GetRemoteData != nil {
		// 3.查询mysql,mongodb,grpc,http api等方式
		item.GetRemoteData(failedList2, idResM)
		b.saveRedisCache(failedList2, idResM, item.RedisExpiration)
	}

	return idResM
}

func (b TertiaryCache) ProcessCache(idList []interface{}, rsp map[interface{}]interface{}) (failed []interface{}) {
	failed = make([]interface{}, 0, len(idList))

	for _, id := range idList {
		v, ok := b.localCache.GetPtr(b.CacheKeyFormatFunc(id))
		if ok {
			rsp[id] = v
		} else {
			failed = append(failed, id)
		}

	}
	return failed
}

func (b TertiaryCache) DelRedisCache(keyList []string) (int, error) {
	pipe := b.redisCmd.Pipeline()
	defer pipe.Close()
	for _, key := range keyList {
		pipe.Del(key)
	}
	_, err := pipe.Exec()
	return len(keyList), err
}

func (b TertiaryCache) RedisCache(idList []interface{}, rsp map[interface{}]interface{}, dataType reflect.Type) (failed, success []interface{}) {
	failed = make([]interface{}, 0, len(idList))
	success = make([]interface{}, 0, len(idList))
	pipe := b.redisCmd.Pipeline()
	defer pipe.Close()

	resCmdM := make(map[interface{}]*redis.StringCmd, len(idList))
	for _, id := range idList {
		cacheKey := b.CacheKeyFormatFunc(id)
		resCmdM[id] = pipe.Get(cacheKey)
	}

	_, err := pipe.Exec()
	util.CheckError(err)

	for id, cmd := range resCmdM {
		if b.Unmarshal != nil {
			bytes, err := cmd.Bytes()
			if err != nil {
				util.CheckError(err)
				failed = append(failed, id)
				continue
			}
			value := reflect.New(dataType).Interface()
			err = b.Unmarshal(bytes, value)
			util.CheckError(err)
			rsp[id] = value
			success = append(success, id)
		} else {
			// 没有序列化方式，统一取出字符串
			value, err := cmd.Result()
			util.CheckError(err)
			if len(value) > 0 {
				rsp[id] = value
				success = append(success, id)
			} else {
				failed = append(failed, id)
			}
		}

	}
	return failed, success
}

// 存数据
func (b TertiaryCache) saveRedisCache(idList []interface{}, rsp map[interface{}]interface{}, expiration time.Duration) {
	if len(idList) == 0 {
		return
	}

	pipe := b.redisCmd.Pipeline()
	defer pipe.Close()

	for _, id := range idList {
		if _, ok := rsp[id]; !ok {
			// 没有查到不写redis
			continue
		}

		cacheKey := b.CacheKeyFormatFunc(id)
		if b.Marshal != nil {
			value, err := b.Marshal(rsp[id])
			util.CheckError(err)
			pipe.Set(cacheKey, value, expiration)
		} else {
			//fmt.Println(cacheKey, rsp[id], expiration)
			pipe.Set(cacheKey, rsp[id], expiration)
		}

	}

	_, err := pipe.Exec()
	util.CheckError(err)
}

func (b TertiaryCache) saveProcessCache(idList []interface{}, rsp map[interface{}]interface{}) {
	if len(idList) == 0 {
		return
	}

	for _, id := range idList {
		cacheKey := b.CacheKeyFormatFunc(id)
		b.localCache.SetPtr(cacheKey, rsp[id])
	}
}

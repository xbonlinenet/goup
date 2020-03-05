package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/xbonlinenet/goup/frame/cache/internal/lrucache"
	"github.com/xbonlinenet/goup/frame/data"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"github.com/vmihailenco/msgpack"
)

var (
	cacheRepo         = make(map[string]*Codec)
	tertiaryCacheRepo = make(map[string]*TertiaryCache, 1)

	lock              = sync.Mutex{}
	tertiaryCacheLock = sync.Mutex{}
)

func GetTertiaryCacheRepo() map[string]*TertiaryCache {
	return tertiaryCacheRepo
}

func internalGetTertiaryCache(name string) *TertiaryCache {
	if c, ok := tertiaryCacheRepo[name]; ok {
		return c
	}

	return nil
}

func GetCacheRepo() map[string]*Codec {
	return cacheRepo
}

func interalGetCacheCodec(name string) *Codec {
	if c, ok := cacheRepo[name]; ok {
		return c
	}
	return nil
}

func BuildCacheCodec(name string,
	redisClient rediser,
	lruEnable bool,
	lruMaxSize int,
	lruExpireDuration time.Duration) *Codec {

	lock.Lock()
	defer lock.Unlock()

	c := interalGetCacheCodec(name)
	if c != nil {
		return c
	}

	codec := &Codec{
		Redis: redisClient,

		Marshal: func(v interface{}) ([]byte, error) {
			return msgpack.Marshal(v)
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return msgpack.Unmarshal(b, v)
		},
	}
	if lruEnable {
		codec.UseLocalCache(lruMaxSize, lruExpireDuration)
	}
	cacheRepo[name] = codec
	return codec
}

func GetCacheCodec(name string) *Codec {
	c := interalGetCacheCodec(name)
	if c != nil {
		return c
	}
	redisName := viper.GetString(fmt.Sprintf("cache.%s.redis", name))
	redisCient := data.MustGetRedis(redisName)
	lruEnable := viper.GetBool(fmt.Sprintf("cache.%s.lru-enable", name))
	lruMaxSize := 1
	lruExpireDuration := 30 * time.Second
	if lruEnable {
		lruMaxSize = viper.GetInt(fmt.Sprintf("cache.%s.lru-max-size", name))
		lruExpireDuration = viper.GetDuration(fmt.Sprintf("cache.%s.lru-expire-duration", name))
	}

	c = BuildCacheCodec(name, redisCient, lruEnable, lruMaxSize, lruExpireDuration)
	return c
}

func newTertiaryCache(name string,
	redisCmd redis.Cmdable,
	lruMaxSize int,
	lruExpireDuration time.Duration) *TertiaryCache {
	tertiaryCacheLock.Lock()
	defer tertiaryCacheLock.Unlock()

	localCache := lrucache.New(lruMaxSize, lruExpireDuration)

	res := &TertiaryCache{
		localCache: localCache,
		redisCmd:   redisCmd,
		Marshal: func(v interface{}) ([]byte, error) {
			return msgpack.Marshal(v)
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return msgpack.Unmarshal(b, v)
		},
	}

	tertiaryCacheRepo[name] = res
	return res
}

func GetTertiaryCache(name string) *TertiaryCache {

	c := internalGetTertiaryCache(name)
	if c != nil {
		return c
	}
	redisName := viper.GetString(fmt.Sprintf("cache.%s.redis", name))
	redisCient := data.MustGetRedis(redisName)

	lruEnable := viper.GetBool(fmt.Sprintf("cache.%s.lru-enable", name))
	lruMaxSize := 1
	lruExpireDuration := 30 * time.Second
	if lruEnable {
		lruMaxSize = viper.GetInt(fmt.Sprintf("cache.%s.lru-max-size", name))
		lruExpireDuration = viper.GetDuration(fmt.Sprintf("cache.%s.lru-expire-duration", name))
	}

	c = newTertiaryCache(name, redisCient, lruMaxSize, lruExpireDuration)
	return c
}

func newTertiaryCacheWithoutMarshal(name string,
	redisCmd redis.Cmdable,
	lruMaxSize int,
	lruExpireDuration time.Duration) *TertiaryCache {
	tertiaryCacheLock.Lock()
	defer tertiaryCacheLock.Unlock()

	localCache := lrucache.New(lruMaxSize, lruExpireDuration)

	res := &TertiaryCache{
		localCache: localCache,
		redisCmd:   redisCmd,
	}

	tertiaryCacheRepo[name] = res
	return res
}

func GetTertiaryCacheWithoutMarshal(name string) *TertiaryCache {

	c := internalGetTertiaryCache(name)
	if c != nil {
		return c
	}
	redisName := viper.GetString(fmt.Sprintf("cache.%s.redis", name))
	redisCient := data.MustGetRedis(redisName)

	lruEnable := viper.GetBool(fmt.Sprintf("cache.%s.lru-enable", name))
	lruMaxSize := 1
	lruExpireDuration := 30 * time.Second
	if lruEnable {
		lruMaxSize = viper.GetInt(fmt.Sprintf("cache.%s.lru-max-size", name))
		lruExpireDuration = viper.GetDuration(fmt.Sprintf("cache.%s.lru-expire-duration", name))
	}

	c = newTertiaryCacheWithoutMarshal(name, redisCient, lruMaxSize, lruExpireDuration)
	return c
}

func NewLoadingCache(lruMaxSize int, lruExpireDuration time.Duration) *LoadingCache {

	localCache := lrucache.New(lruMaxSize, lruExpireDuration)

	return &LoadingCache{localCache: localCache}
}

type LoadingCache struct {
	// lru
	localCache *lrucache.Cache
}

func (c *LoadingCache) Once(key string, f func() interface{}) interface{} {
	v, ok := c.localCache.GetPtr(key)
	if ok {
		return v
	}

	v = f()
	if v == nil {
		return nil
	}
	c.localCache.SetPtr(key, v)
	return v
}

package data

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"

	"github.com/xbonlinenet/goup/frame/util"
)

var redisMgr *RedisMgr

var (
	// ErrRedisConfig 配置错误
	ErrRedisConfig = errors.New("redis config error")
	// ErrRedisMgrNotInited Redis Manager 还未初始化
	ErrRedisMgrNotInited = errors.New("redis manager not inited")
)

// ErrRedisConfigNotFound 配置未找到错误
type ErrRedisConfigNotFound struct {
	ConfigName string
}

func (e ErrRedisConfigNotFound) Error() string {
	return fmt.Sprintf("redis config (%s) not found", e.ConfigName)
}

// ErrRedisInitError 初始化 Redis 错误
type ErrRedisInitError struct {
	Err error
}

func (e ErrRedisInitError) Error() string {
	return fmt.Sprintf("init redis error %s", e.Err.Error())
}

// InitRedisMgr 初始化 Redis
func InitRedisMgr() {
	redisMgr = newRedisMgr(viper.Sub("data.redis"))
}

// UninitRedisMgr 反初始化 Redis 相关
func UninitRedisMgr() {
	if redisMgr != nil {
		redisMgr.Close()
		redisMgr = nil
	}
}

// GetRedis 获取 redis
func GetRedis(name string) (redis.Cmdable, error) {
	if redisMgr == nil {
		panic(ErrRedisMgrNotInited)
	}

	return redisMgr.getRedis(name)
}

// MustGetRedis 获取 redis，如果获取失败，直接报错
func MustGetRedis(name string) redis.Cmdable {
	if redisMgr == nil {
		panic(ErrRedisMgrNotInited)
	}

	return redisMgr.mustGetRedis(name)
}

// newRedisMgr 根据配置创建 RedisMgr
func newRedisMgr(conf *viper.Viper) *RedisMgr {
	redisMgr := &RedisMgr{
		redisMap:    make(map[string]*redis.Client),
		clusterMap:  make(map[string]*redis.ClusterClient),
		mutex:       &sync.Mutex{},
		redisConfig: conf,
	}
	return redisMgr
}

// RedisMgr Redis管理器
type RedisMgr struct {
	redisMap    map[string]*redis.Client
	clusterMap  map[string]*redis.ClusterClient
	mutex       *sync.Mutex
	redisConfig *viper.Viper
}

// GetRedis 获取 redis 实例
func (mgr *RedisMgr) getRedis(name string) (redis.Cmdable, error) {
	config := mgr.redisConfig.Sub(name)
	if config == nil {
		return nil, ErrRedisConfigNotFound{name}
	}

	isCluster := config.GetBool("cluster")
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	if isCluster {
		cluster, ok := mgr.clusterMap[name]
		if ok {
			return cluster, nil
		}

		cluster, err := initRedisClusterClient(config)
		if err != nil {
			return nil, err
		}
		mgr.clusterMap[name] = cluster
		return cluster, nil
	}

	client, ok := mgr.redisMap[name]
	if ok {
		return client, nil
	}

	client, err := initRedisClient(config)
	if err != nil {
		return nil, err
	}
	mgr.redisMap[name] = client
	return client, nil

}

func (mgr *RedisMgr) mustGetRedis(name string) redis.Cmdable {
	config := mgr.redisConfig.Sub(name)
	if config == nil {
		panic(ErrRedisConfigNotFound{name})
	}

	isCluster := config.GetBool("cluster")
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	if isCluster {
		cluster, ok := mgr.clusterMap[name]
		if ok {
			return cluster
		}

		cluster, err := initRedisClusterClient(config)
		util.CheckError(err)

		mgr.clusterMap[name] = cluster
		return cluster
	}
	client, ok := mgr.redisMap[name]
	if ok {
		return client
	}

	client, err := initRedisClient(config)
	if err != nil {
		panic(ErrRedisInitError{err})
	}

	mgr.redisMap[name] = client
	return client

}

func initRedisClient(config *viper.Viper) (*redis.Client, error) {
	addr := config.GetString("addr")
	if len(addr) == 0 {
		addr = "localhost:6379"
	}
	password := config.GetString("password")
	poolSize := config.GetInt("pool-size")
	db := config.GetInt("db")

	client := redis.NewClient(&redis.Options{
		Addr:               addr,
		Password:           password,
		PoolSize:           poolSize,
		DB:                 db,
		IdleTimeout:        50 * time.Second,
		IdleCheckFrequency: 20 * time.Second,
	})
	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func initRedisClusterClient(config *viper.Viper) (*redis.ClusterClient, error) {
	addrs := config.GetStringSlice("addrs")
	if len(addrs) <= 0 {
		return nil, ErrRedisInitError{errors.New("cluster hasn't any addr")}
	}
	password := config.GetString("password")
	poolSize := config.GetInt("pool-size")

	cluster := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        addrs,
		Password:     password,
		PoolSize:     poolSize,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolTimeout:  30 * time.Second,
	})

	_, err := cluster.Ping().Result()
	if err != nil {
		return nil, ErrRedisInitError{err}
	}

	return cluster, nil
}

// Close 关闭
func (mgr *RedisMgr) Close() {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	for _, client := range mgr.redisMap {
		client.Close()
	}

	for _, cluster := range mgr.clusterMap {
		cluster.Close()
	}

	mgr.redisMap = make(map[string]*redis.Client)
	mgr.clusterMap = make(map[string]*redis.ClusterClient)
}

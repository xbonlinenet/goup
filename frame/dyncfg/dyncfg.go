package dyncfg

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/deckarep/golang-set"

	"github.com/xbonlinenet/goup/frame/alter"
	"github.com/xbonlinenet/goup/frame/data"
	"github.com/xbonlinenet/goup/frame/log"
	simplejson "github.com/bitly/go-simplejson"
	"github.com/go-redis/redis"
)

const (
	configPrefix = "AppConf"
)

var (
	config      = make(map[string]string)
	redisClient redis.Cmdable
	app         string
)

var (
	ErrCfgNotFound = errors.New("DynCfg Cfg not found")
)

func Init(appName string, client string, ctx context.Context) {
	app = appName
	redisClient = data.MustGetRedis(client)
	updateAllConfig()

	go loopUpdateConfig(ctx)
}

// 遍历加载 redis 中的数据
// 可以考虑优化 使用 redis 的订阅增量更新
func loopUpdateConfig(ctx context.Context) {
loop:
	for {
		select {
		case <-time.After(10 * time.Second):
			updateAllConfig()
		case <-ctx.Done():
			break loop
		}
	}
	log.Sugar().Infof("exit update config")
}

func updateAllConfig() {
	key := fmt.Sprintf("%s.%s", configPrefix, app)
	log.Sugar().Infof("updateAllConfig, key: %s", key)

	data, err := redisClient.HGetAll(key).Result()
	if err == redis.Nil {
		return
	}
	if err != nil {
		alter.NotifyError("updateAllConfig failed", err)
	}
	log.Sugar().Debugf("All Config: %v", data)
	config = data
}

// MustGetString
// 获取字符串配置，如果不存在或者为空字符串则会抛出异常
func MustGetString(item string) string {
	if v, ok := config[item]; ok {
		if len(v) == 0 {
			panic(ErrCfgNotFound)
		}
		return v
	}
	panic(ErrCfgNotFound)
}

func GetString(item string, defaultValue string) string {
	if v, ok := config[item]; ok {
		return v
	}
	return defaultValue
}

func MustGetFloat64(item string) float64 {
	if v, ok := config[item]; ok {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			panic(err)
		}
		return f
	}
	panic(ErrCfgNotFound)
}

func GetFloat64(item string, defaultValue float64) float64 {
	if v, ok := config[item]; ok {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return defaultValue
		}
		return f
	}
	return defaultValue
}

func MustGetStringSlice(item string) []string {
	if v, ok := config[item]; ok {
		js, err := simplejson.NewJson([]byte(v))
		if err != nil {
			panic(err)
		}
		arr, err := js.Array()
		if err != nil {
			panic(err)
		}

		result := make([]string, 0, len(arr))
		for _, item := range arr {
			str := item.(string)
			result = append(result, str)
		}
		return result
	}
	panic(ErrCfgNotFound)
}

func GetSet(item string, defaultValue mapset.Set) mapset.Set {
	if v, ok := config[item]; ok {
		js, err := simplejson.NewJson([]byte(v))
		if err != nil {
			panic(err)
		}
		arr, err := js.Array()
		if err != nil {
			panic(err)
		}
		return mapset.NewSetFromSlice(arr)
	}
	return defaultValue
}
func MustGetSet(item string) mapset.Set {
	if v, ok := config[item]; ok {
		js, err := simplejson.NewJson([]byte(v))
		if err != nil {
			panic(err)
		}
		arr, err := js.Array()
		if err != nil {
			panic(err)
		}
		return mapset.NewSetFromSlice(arr)
	}
	panic(ErrCfgNotFound)
}

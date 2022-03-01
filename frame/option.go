package frame

import (
	"github.com/gin-gonic/gin"
	"github.com/xbonlinenet/goup/frame/data"
)

// An Option configures
type Option interface {
	apply(*bootstarpServerConfig)
}

// optionFunc wraps a func so it satisfies the Option interface.
type optionFunc func(*bootstarpServerConfig)

func (f optionFunc) apply(log *bootstarpServerConfig) {
	f(log)
}

// BeforeInit 在框架初始化之前运行
func BeforeInit(f func()) Option {
	return optionFunc(func(cfg *bootstarpServerConfig) {
		cfg.beforeInit = f
	})
}

// BeforeServerRun 在web服务启动之前运行
func BeforeServerRun(f func()) Option {
	return optionFunc(func(cfg *bootstarpServerConfig) {
		cfg.beforeServerRun = f
	})
}

// CustomRouter 自定义路由，用于突破框架的 json api 的局限性
func CustomRouter(f func(r *gin.Engine)) Option {
	return optionFunc(func(cfg *bootstarpServerConfig) {
		cfg.customRouter = f
	})
}

// Version 自定义 /version path 返回内容
func Version(f func(c *gin.Context)) Option {
	return optionFunc(func(cfg *bootstarpServerConfig) {
		cfg.versionHandler = f
	})
}

// ReportApi 上报接口到文档中心
func ReportApi(addr string) Option {
	return optionFunc(func(cfg *bootstarpServerConfig) {
		cfg.reportApiDocAddr = addr
	})
}

func DisableDbInit() Option {
	return optionFunc(func(cfg *bootstarpServerConfig) {
		cfg.initDbDisabled = true
	})
}

func DisableRedisInit() Option {
	return optionFunc(func(cfg *bootstarpServerConfig) {
		cfg.initRedisDisabled = true
	})
}

func DisableEsInit() Option {
	return optionFunc(func(cfg *bootstarpServerConfig) {
		cfg.initEsDisabled = true
	})
}

func DisableKafkaInit() Option {
	return optionFunc(func(cfg *bootstarpServerConfig) {
		cfg.initKafkaDisabled = true
	})
}

func EnableHttpHealthz() Option {
	return optionFunc(func(cfg *bootstarpServerConfig) {
		cfg.enableHttpHealthz = true
	})
}

// Middlewares 添加自定义的 middleware
func Middlewares(list []gin.HandlerFunc) Option {
	return optionFunc(func(cfg *bootstarpServerConfig) {
		cfg.middlewareList = list
	})
}

// CustomSqlConf 设置自定义的 SQL 实例
func CustomSqlConf(conf map[string]*data.SQLConfig) Option {
	return optionFunc(func(cfg *bootstarpServerConfig) {
		cfg.customSqlConf = conf
	})
}

func CustomRedisConf(conf map[string]*data.RedisConfig) Option {
	return optionFunc(func(cfg *bootstarpServerConfig) {
		cfg.custonRedisConf = conf
	})
}

func SpecifyApiPathPrefix(apiPrefix string) Option {
	return optionFunc(func(cfg *bootstarpServerConfig) {
		cfg.customApiPathPrefix = apiPrefix
	})
}

// RegisterHook 注册hook函数，在服务优雅关闭之前执行
func RegisterHook(f func()) Option {
	return optionFunc(func(cfg *bootstarpServerConfig) {
		cfg.Hook = f
	})
}

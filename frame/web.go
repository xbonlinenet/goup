package frame

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/xbonlinenet/goup/frame/gateway"
	"github.com/xbonlinenet/goup/frame/log"
	"github.com/xbonlinenet/goup/frame/recovery"
	"github.com/xbonlinenet/goup/frame/util"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

func BootstrapServer(ctx context.Context, options ...Option) {

	config := &bootstarpServerConfig{}

	for _, opt := range options {
		opt.apply(config)
	}

	if config.beforeInit != nil {
		config.beforeInit()
	}

	InitFramework()
	defer UnInitFramework()

	r := gin.New()

	p := ginprometheus.NewPrometheus("gin")
	p.ReqCntURLLabelMappingFn = func(c *gin.Context) string {
		url := c.Request.URL.Path
		return url
	}

	p.Use(r)

	debug := viper.GetBool("application.debug")
	if debug {
		gin.SetMode("debug")
		r.GET("/doc/list", gateway.ApiList)
		r.GET("/doc/detail", gateway.ApiDetail)
	} else {
		gin.SetMode("release")
	}

	r.Use(recovery.Recovery())
	r.Use(gateway.APIMiddleware())

	if config.customRouter != nil {
		config.customRouter(r)
	}

	if config.versionHandler != nil {
		r.GET("/version", config.versionHandler)
	}

	if config.beforeServerRun != nil {
		config.beforeServerRun()
	}

	addr := viper.GetString("server.addr")
	err := r.Run(addr)
	util.CheckError(err)
	log.Sugar().Errorf("Occur err: %s", err.Error())
}

type bootstarpServerConfig struct {
	beforeInit      func()
	beforeServerRun func()
	customRouter    func(r *gin.Engine)
	versionHandler  func(c *gin.Context)
}

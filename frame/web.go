package frame

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/xbonlinenet/goup/frame/gateway"
	"github.com/xbonlinenet/goup/frame/log"
	"github.com/xbonlinenet/goup/frame/recovery"
	"github.com/xbonlinenet/goup/frame/util"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

func BootstrapServer(
	beforeInit func(),
	berforeServerRun func(),
	customRouter func(r *gin.Engine),
	version func(c *gin.Context)) {

	InitFramework()
	defer UnInitFramework()

	if beforeInit != nil {
		beforeInit()
	}

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

	r.GET("/version", version)

	if berforeServerRun != nil {
		berforeServerRun()
	}

	addr := viper.GetString("server.addr")
	err := r.Run(addr)
	util.CheckError(err)
	log.Sugar().Errorf("Occur err: %s", err.Error())
}

package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/xbonlinenet/goup/frame"
	"github.com/xbonlinenet/goup/frame/log"
	"github.com/xbonlinenet/goup/frame/recovery"
	"github.com/xbonlinenet/goup/frame/util"
	"github.com/xbonlinenet/goup/frame/web"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

func main() {
	frame.Bootstrap(run)
}

func run() {

	go func() {
		http.ListenAndServe("0.0.0.0:6060", nil)
	}()

	r := gin.New()
	// add prometheus middlewares
	p := ginprometheus.NewPrometheus("gin")
	p.Use(r)

	debug := viper.GetBool("application.debug")
	if debug {
		gin.SetMode("debug")
	} else {
		gin.SetMode("release")
	}
	r.Use(recovery.Recovery())
	r.GET("/version", web.Version)
	registerRouter(r.Group("api"))

	addr := viper.GetString("server.addr")

	err := r.Run(addr)
	util.CheckError(err)
	log.Sugar().Errorf("Occur err: %s", err.Error())
}

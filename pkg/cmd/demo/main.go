package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"

	"github.com/gin-gonic/gin"
	"github.com/xbonlinenet/goup/frame"
	"github.com/xbonlinenet/goup/frame/flags"
)

func version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":    flags.GitCommit,
		"build_time": flags.BuildTime,
		"build_env":  flags.BuildEnv,
	})
}

func customRouter(r *gin.Engine) {
	r.GET("hello", func(c *gin.Context) {
		c.String(200, "Hello world!")
	})
}

func main() {
	ctx := context.Background()
	frame.BootstrapServer(
		ctx,
		frame.BeforeServerRun(registerRouter),
		frame.Version(version),
		frame.CustomRouter(customRouter),
		frame.ReportApi("http://192.168.0.22:14000/api/doc/report"),
	)
}

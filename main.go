package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/xbonlinenet/goup/demo"
	"github.com/xbonlinenet/goup/frame"
	"github.com/xbonlinenet/goup/frame/flags"
	"github.com/xbonlinenet/goup/frame/gateway"
	"github.com/xbonlinenet/goup/frame/log"
)

func main() {
	ctx := context.Background()
	frame.BootstrapServer(
		ctx,
		frame.BeforeServerRun(registerRouter),
		frame.Version(version),
		frame.CustomRouter(customRouter),
		frame.ReportApi("http://192.168.0.22:14000/api/doc/report"),
		frame.EnableconfigCenterDefalutLocalCacheDir(true),
	)
}

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

func registerRouter() {
	ilikeCORSHandler := gateway.NewCORSHandler([]string{"www.ilikee.vn", "ilikee.vn", "0.0.0.0:8000"})

	gateway.RegisterAPI("demo", "echo", "Demo for echo", demo.EchoHandler{})
	gateway.RegisterAPI("demo", "cors_echo", "Demo for cors",
		demo.EchoHandler{},
		gateway.WithCORSHandler(ilikeCORSHandler),
	)
	gateway.RegisterAPI("demo", "redis", "Demo for reids incr", demo.RedisHandler{}, gateway.ExtInfo(map[string]string{"test": "<a href='https://www.null.com'>test</a>"}))
	gateway.RegisterAPI("demo", "mysql", "Demo for mysql ", demo.MysqlHandler{})
	gateway.RegisterAPI("demo", "config", "Demo for config center ", demo.ConfigHandler{})
	gateway.RegisterAPI("demo", "pre", "Demo for pre handler, normally used in login filter",
		demo.PreHandler{},
		gateway.HandlerFunc(demoPreHandler),
		gateway.HandlerFunc(demoPreHandler2),
		gateway.HandlerFunc(demoPreHandler3),
	)
	gateway.RegisterAPI("demo", "sleep", "Demo for sleep", demo.SleepHandler{})
	gateway.RegisterAPI("demo", "doc", "Demo complex sturct", demo.DocHandler{})
}

func demoPreHandler(c *gin.Context, ctx *gateway.ApiContext) *gateway.Resp {
	log.Default().Info("into demoPreHandler")

	ctx.Keys["message"] = "This has handled by demoPreHandler"
	return nil
}

func demoPreHandler2(c *gin.Context, ctx *gateway.ApiContext) *gateway.Resp {
	log.Default().Info("into demoPreHandler2")

	if _, ok := ctx.Keys["message"]; ok {
		message := ctx.Keys["message"].(string)
		newMsg := fmt.Sprintf("%s\nalso has handled by demoPreHandler2", message)
		ctx.Keys["message"] = newMsg
	}

	return nil
}

func demoPreHandler3(c *gin.Context, ctx *gateway.ApiContext) *gateway.Resp {
	log.Default().Info("into demoPreHandler3")

	if _, ok := ctx.Keys["message"]; ok {
		message := ctx.Keys["message"].(string)
		newMsg := fmt.Sprintf("%s\nhandled by demoPreHandler3 too", message)
		ctx.Keys["message"] = newMsg
	}

	return nil
}

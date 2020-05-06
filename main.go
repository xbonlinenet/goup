package main

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/xbonlinenet/goup/demo"
	"github.com/xbonlinenet/goup/frame"
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
	)
}

func version(c *gin.Context) {

}

func customRouter(r *gin.Engine) {
	r.GET("hello", func(c *gin.Context) {
		c.String(200, "Hello world!")
	})
}

func registerRouter() {
	gateway.RegisterAPI("demo", "echo", "Demo for echo", demo.EchoHandler{})
	gateway.RegisterAPI("demo", "redis", "Demo for reids incr", demo.RedisHandler{})
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

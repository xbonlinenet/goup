package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/xbonlinenet/goup/demo"
	"github.com/xbonlinenet/goup/frame"
	"github.com/xbonlinenet/goup/frame/gateway"
)

func main() {
	ctx := context.Background()
	frame.BootstrapServer(ctx, frame.BeforeServerRun(registerRounter), frame.Version(version), frame.CustomRouter(customRounter), frame.ReportApi("http://192.168.0.22:14000/api/doc/report"))
}

func version(c *gin.Context) {

}

func customRounter(r *gin.Engine) {
	r.GET("hello", func(c *gin.Context) {
		c.String(200, "Hello world!")
	})
}

func registerRounter() {
	gateway.RegisterAPI("demo", "echo", "Demo for echo", demo.EchoHandler{})
	gateway.RegisterAPI("demo", "redis", "Demo for reids incr", demo.RedisHandler{})
	gateway.RegisterAPI("demo", "mysql", "Demo for mysql ", demo.MysqlHandler{})
	gateway.RegisterAPI("demo", "config", "Demo for config center ", demo.ConfigHandler{})
	gateway.RegisterAPI("demo", "pre", "Demo for pre handler, normally used in login filter", demo.PreHandler{}, gateway.HandlerFunc(demoPreHandler))
	gateway.RegisterAPI("demo", "sleep", "Demo for sleep", demo.SleepHandler{})
}

func demoPreHandler(c *gin.Context, ctx *gateway.ApiContext) *gateway.Resp {
	ctx.Keys["message"] = "This has handled by demoPreHandler"
	return nil
}

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/xbonlinenet/goup/demo"
	"github.com/xbonlinenet/goup/frame"
	"github.com/xbonlinenet/goup/frame/gateway"
)

func main() {
	frame.BootstrapServer(nil, registerRounter, customRounter, version)
}

func version(c *gin.Context) {

}

func customRounter(r *gin.Engine) {
	r.GET("hello", func(c *gin.Context) {
		c.String(200, "Hello world!")
	})
}

func registerRounter() {
	gateway.RegisterAPI("demo", "echo", "Echo interface for debug", demo.EchoHandler{})
}

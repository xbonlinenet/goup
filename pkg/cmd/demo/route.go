package main

import (
	"github.com/gin-gonic/gin"
	"github.com/xbonlinenet/goup/pkg/cmd/demo/api"
)

func registerRouter(r *gin.RouterGroup) {
	r.POST("test", api.Demo)

}

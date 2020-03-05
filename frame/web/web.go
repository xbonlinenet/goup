package web

import (
	"net/http"

	"github.com/xbonlinenet/goup/frame/util"
	"github.com/gin-gonic/gin"
)

func Version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": util.Version, "compile": util.Compile})
}

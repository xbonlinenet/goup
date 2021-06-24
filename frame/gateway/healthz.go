package gateway

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func HttpHealthz(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("I'm running now"))

}
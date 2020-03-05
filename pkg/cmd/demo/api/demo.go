package api

import "github.com/gin-gonic/gin"

// Demo demo
func Demo(c *gin.Context) {
	c.JSON(200, map[string]string{"message": "hello"})
}

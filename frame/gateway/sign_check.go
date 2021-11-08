package gateway

import (
	"github.com/gin-gonic/gin"
)

//签名校验
type SignCheckHandler struct {
	signCheck func(c *gin.Context) bool
}

func NewSignCheckHandler(signCheck func(c *gin.Context) bool) *SignCheckHandler {
	return &SignCheckHandler{
		signCheck: signCheck,
	}
}

func (h *SignCheckHandler) CheckSign(c *gin.Context) bool {
	return h.signCheck(c)
}

//注册
func WithSignCheckHandler(h *SignCheckHandler) Option {
	return optionFunc(func(handler *HandlerInfo) {
		handler.signCheckHandler = h
	})
}

package gateway

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

// 签名校验
type SignCheckHandler struct {
	signCheck func(c *gin.Context) bool
}

func NewSignCheckHandler(signCheck func(c *gin.Context) bool) *SignCheckHandler {

	return &SignCheckHandler{
		signCheck: signCheck,
	}
}

func (h *SignCheckHandler) CheckSign(c *gin.Context) bool {
	// c.Request.Header
	return h.signCheck(c)
}

// 注册
func WithSignCheckHandler(h *SignCheckHandler) Option {
	return optionFunc(func(handler *HandlerInfo) {
		handler.signCheckHandler = h
	})
}

type SignCheckHandlerV2 struct {
	signCheck func(u *url.URL, headers http.Header, body []byte) bool
}

func (h *SignCheckHandlerV2) CheckSign(u *url.URL, headers http.Header, body []byte) bool {
	// c.Request.Header
	return h.signCheck(u, headers, body)
}

func NewSignCheckHandlerV2(signCheck func(u *url.URL, headers http.Header, body []byte) bool) *SignCheckHandlerV2 {
	return &SignCheckHandlerV2{
		signCheck: signCheck,
	}
}

func WithSignCheckHandlerV2(h *SignCheckHandlerV2) Option {
	return optionFunc(func(handler *HandlerInfo) {
		handler.signCheckHandlerV2 = h
	})
}

package gateway

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xbonlinenet/goup/frame/log"
)

type CORSHandler struct {
	AllowHosts []string
}

func NewCORSHandler(allowHosts []string) *CORSHandler {
	return &CORSHandler{
		AllowHosts: allowHosts,
	}
}

func (h *CORSHandler) CheckOriginByRequest(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	return h.CheckOrigin(origin)
}

func (h *CORSHandler) CheckOrigin(origin string) bool {
	if origin == "" {
		// 非浏览器请求
		return false
	}

	originURL, err := url.Parse(origin)
	if err != nil {
		log.Default().Warn("ParseOriginErr", zap.Error(err))
		return false
	}

	var isAllow bool
	for _, allowHost := range h.AllowHosts {
		if originURL.Host == allowHost {
			isAllow = true
			break
		}
	}

	return isAllow
}

func (h *CORSHandler) BuildCORSHeaders(r *http.Request) map[string]string {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}

	headers := map[string]string{
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Methods":     "GET,HEAD,POST,PUT,DELETE",
		"Access-Control-Allow-Origin":      origin,
	}
	return headers
}

func (h *CORSHandler) WriteCORSHeader(ctx *gin.Context) {
	headers := h.BuildCORSHeaders(ctx.Request)
	for key, val := range headers {
		ctx.Header(key, val)
	}
}

func WithCORSHandler(h *CORSHandler) Option {
	return optionFunc(func(handler *HandlerInfo) {
		handler.corsHandler = h
	})
}

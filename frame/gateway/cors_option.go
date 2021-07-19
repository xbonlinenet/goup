package gateway

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xbonlinenet/goup/frame/util"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strings"

	"github.com/xbonlinenet/goup/frame/log"
)

type CORSHandler struct {
	AllowHosts []string
	AllowAll   bool
	allowedCustomHeaders []string
}

func NewCORSHandler(allowHosts []string, customHeaders ...string) *CORSHandler {
	var allowAll bool

	for _, host := range allowHosts {
		if host == "*" {
			allowAll = true
			break
		}
	}

	fmt.Println("NewCORSHandler called, customHeaders:", customHeaders)

	return &CORSHandler{
		AllowHosts: allowHosts,
		AllowAll:   allowAll,
		allowedCustomHeaders: customHeaders,
	}
}

func (h *CORSHandler) CheckOriginByRequest(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	return h.CheckOrigin(origin)
}

func (h *CORSHandler) CheckOrigin(origin string) bool {
	if origin == "" {
		// 非浏览器请求
		log.Default().Warn("CheckOrigin(): found 'Origin' is empty from header!")
		return false
	}

	if h.AllowAll {
		return true
	}

	originURL, err := url.Parse(origin)
	if err != nil {
		log.Default().Warn("ParseOriginErr", zap.Error(err), zap.String("origin", origin))
		return false
	}

	var isAllow bool
	for _, allowHost := range h.AllowHosts {
		if originURL.Host == allowHost {
			isAllow = true
			break
		}
	}
	log.Default().Debug("CheckOrigin() result", zap.Bool("isAllow", isAllow),
		zap.String("origin", origin))

	return isAllow
}

func (h *CORSHandler) BuildCORSHeaders(r *http.Request) map[string]string {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}

	allowedHeaders := []string{
		"Origin",
		"X-Requested-With",
		"Content-Type",
		"Accept",
	}
	if len(h.allowedCustomHeaders) > 0 {
		for _, customHeader := range h.allowedCustomHeaders {
			if !util.StringArrayContains(allowedHeaders, customHeader) {
				allowedHeaders = append(allowedHeaders, customHeader)
			}
		}
	}
	allowedHeadersStr := strings.Join(allowedHeaders, ",")
	log.Default().Debug("cors allowedHeaders", zap.String("headers", allowedHeadersStr),
		zap.Any("customHeaders", h.allowedCustomHeaders))

	headers := map[string]string{
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Methods":     "GET,HEAD,POST,PUT,DELETE",
		"Access-Control-Allow-Origin":      origin,
		"Access-Control-Allow-Headers":     allowedHeadersStr,
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

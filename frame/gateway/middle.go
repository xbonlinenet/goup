package gateway

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/xbonlinenet/goup/frame/alter"
	"github.com/xbonlinenet/goup/frame/log"
	"github.com/xbonlinenet/goup/frame/recovery"
	"go.uber.org/zap"
)

var (
	apiHandlerFuncMap = map[string]*HandlerInfo{}
)

func init() {
	prometheus.MustRegister(requestLatency)
	prometheus.MustRegister(invalidRequestCounter)

}

// requestLatency 接口延迟
var requestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: "request_latency",
	Help: "stat request latency by seconds",
}, []string{"api"})

// invalidRequestCounter 统计错误请求数量情况
var invalidRequestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "invalid_request",
	Help: "invalid request by path and code",
}, []string{"path", "code"})

// APIMiddleware 接口中间层
func APIMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if i := strings.Index(c.Request.URL.Path, "/api/"); i == 0 {
			handlerApiRequest(c)
		} else {
			c.Next()
		}
	}
}

func abs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}

func handlerApiRequest(c *gin.Context) {

	start := time.Now()

	defer func() {
		if err := recover(); err != nil {
			stack := recovery.Stack(3)
			log.GetLogger("error").Sugar().Errorf("[Recovery] %s, %v\n %s", err, c.Request.URL.Path, stack)
			alter.Notify(fmt.Sprintf("Error: %s", c.Request.URL.Path), string(stack), c.Request.URL.Path)
			invalidRequestCounter.WithLabelValues(c.Request.URL.Path, strconv.Itoa(ErrUnknowError)).Inc()
			c.AbortWithStatus(500)
		}
	}()

	// 判断是否有路由可以处理
	apiKey := getAPIKey(c.Request.URL.Path)
	if _, ok := apiHandlerFuncMap[apiKey]; !ok {
		c.String(http.StatusNotFound, fmt.Sprintf("api: %s not registered!", apiKey))
		return
	}

	apiContext := &ApiContext{}
	apiContext.Host = c.Request.Host
	apiContext.ClientIP = c.ClientIP()
	apiContext.UserAgent = c.Request.UserAgent()

	// 验证请求是否在有效时间内
	apiHandlerInfo := apiHandlerFuncMap[apiKey]

	defer requestLatency.WithLabelValues(apiKey).Observe(time.Since(start).Seconds())
	// 处理请求
	request := reflect.New(apiHandlerInfo.reqType).Interface()
	err := c.BindJSON(request)
	if err != nil {
		failHanler(c, http.StatusUnauthorized, ErrInvalidParam, err.Error())
		return
	}

	handler := reflect.New(apiHandlerInfo.handler).Interface()
	// 设置请求且信息
	reflect.ValueOf(handler).Elem().FieldByName("Request").Set(reflect.ValueOf(request).Elem())

	if c.GetHeader("Mock") == "true" {
		ret := reflect.ValueOf(handler).Elem().MethodByName("Mock").Call([]reflect.Value{})
		c.PureJSON(200, ret[0].Interface())
	} else {
		ret := reflect.ValueOf(handler).Elem().MethodByName("Handler").Call([]reflect.Value{reflect.ValueOf(apiContext)})
		var response interface{}
		if ret[1].IsNil() {
			response = ret[0].Interface()
		} else {
			err = ret[1].Interface().(error)
			response = Resp{Code: ErrLogicError, Message: err.Error()}
		}

		// access 日志处理
		log.GetLogger("access").Info("api", zap.String("api", apiKey), zap.Any("request", request), zap.Any("context", apiContext), zap.Any("Response", response))
		c.PureJSON(200, response)
	}
}

func failHanler(c *gin.Context, status int, code int, message string) {
	c.JSON(status, Resp{Code: code, Message: message})
	invalidRequestCounter.WithLabelValues(c.Request.URL.Path, strconv.Itoa(code)).Inc()
	log.GetLogger("access_error").Info(c.Request.URL.Path, zap.Int("code", code), zap.String("message", message), zap.String("url", c.Request.URL.String()))
}

func getAPIKey(path string) string {
	api := strings.ReplaceAll(path[5:], "/", ".")
	return api
}

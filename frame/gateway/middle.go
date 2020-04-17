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
	"github.com/xbonlinenet/goup/frame/util"
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
			wrapRequest(handlerApiRequest, c)
			// handlerApiRequest(c)
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
			failHanler(c, http.StatusInternalServerError, ErrUnknowError, string(stack))
		}
	}()

	// 判断是否有路由可以处理
	apiKey := getAPIKey(c.Request.URL.Path)
	if _, ok := apiHandlerFuncMap[apiKey]; !ok {
		c.String(http.StatusNotFound, fmt.Sprintf("api: %s not registered!", apiKey))
		return
	}

	reqId, level, er := getReqInfo(c)
	util.CheckError(er)

	apiContext := new(ApiContext)
	apiContext.ClientIP = c.ClientIP()
	apiContext.Request = c.Request

	// 请求追踪
	apiContext.ReqId = reqId
	apiContext.ReqLevel = level + 1

	// 验证请求是否在有效时间内
	apiHandlerInfo := apiHandlerFuncMap[apiKey]

	// prehandler之前设置
	apiContext.APIConfig.Expires = apiHandlerInfo.expire

	if apiHandlerInfo.preHandler != nil {
		apiContext.Keys = make(map[string]interface{}, 4)
		resp := apiHandlerInfo.preHandler(c, apiContext)
		if resp != nil && resp.Code != 0 {
			failHanler(c, http.StatusOK, resp.Code, resp.Message)
			return
		}
	}

	defer requestLatency.WithLabelValues(apiKey).Observe(time.Since(start).Seconds())
	// 处理请求
	request := reflect.New(apiHandlerInfo.reqType).Interface()

	var err error
	if apiHandlerInfo.pt == formType {
		err = c.ShouldBindQuery(request)
	} else {
		err = c.BindJSON(request)
	}

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
	c.AbortWithStatusJSON(status, Resp{Code: code, Message: message})
	invalidRequestCounter.WithLabelValues(c.Request.URL.Path, strconv.Itoa(code)).Inc()
	log.GetLogger("access_error").Info(c.Request.URL.Path, zap.Int("code", code), zap.String("message", message), zap.String("url", c.Request.URL.String()))
}

func getAPIKey(path string) string {
	api := strings.ReplaceAll(path[5:], "/", ".")
	return api
}

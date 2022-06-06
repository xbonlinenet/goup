package gateway

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
	"go.uber.org/zap"

	"github.com/xbonlinenet/goup/frame/alter"
	"github.com/xbonlinenet/goup/frame/log"
	"github.com/xbonlinenet/goup/frame/perf"
	"github.com/xbonlinenet/goup/frame/recovery"
	"github.com/xbonlinenet/goup/frame/util"
)

var (
	apiHandlerFuncMap = map[string]*HandlerInfo{}
)

const (
	kDefaultApiPathPrefix    = "/api/"
	kAnyApiPathPrefixAllowed = "*"
)

func init() {
	prometheus.MustRegister(requestLatency)
	prometheus.MustRegister(invalidRequestCounter)

}

// requestLatency 接口延迟
var requestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "request_latency",
	Help:    "stat request latency by seconds",
	Buckets: []float64{.005, .01, .025, .05, .1, .15, .2, .25, .5, 1, 2.5, 5, 10},
}, []string{"api"})

// invalidRequestCounter 统计错误请求数量情况
var invalidRequestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "invalid_request",
	Help: "invalid request by path and code",
}, []string{"path", "code"})

// APIMiddleware 接口中间层
func APIMiddleware(customApiPathPrefix string) gin.HandlerFunc {
	if customApiPathPrefix == "" {
		customApiPathPrefix = kDefaultApiPathPrefix
	} else if customApiPathPrefix != kAnyApiPathPrefixAllowed {
		// make sure customApiPathPrefix is like '/xxxx/'
		if !strings.HasPrefix(customApiPathPrefix, "/") {
			customApiPathPrefix = "/" + customApiPathPrefix
		}
		if !strings.HasSuffix(customApiPathPrefix, "/") {
			customApiPathPrefix = customApiPathPrefix + "/"
		}
	}

	return func(c *gin.Context) {
		if customApiPathPrefix == kAnyApiPathPrefixAllowed ||
			strings.HasPrefix(c.Request.URL.Path, customApiPathPrefix) {
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
			notifyMsg, notifyDetail, notifyErrorID := fmt.Sprintf("Error: %s", err), string(stack), c.Request.URL.Path
			alter.Notify(notifyMsg, notifyDetail, notifyErrorID)
			sentry.CaptureMessage(fmt.Sprintf("[Recovery] %s, %v\n%s", err, c.Request.URL.Path, string(stack)))
			failHandler(c, http.StatusInternalServerError, ErrUnknowError, notifyMsg+"\n"+string(stack))
		}
	}()

	// 判断是否有路由可以处理
	apiKey := getAPIKey(c.Request.URL.Path)
	apiHandlerInfo, ok := apiHandlerFuncMap[apiKey]
	if !ok {
		c.Next()
		// c.String(http.StatusNotFound, fmt.Sprintf("api: %s not registered!", apiKey))
		return
	}
	//处理签名校验
	if apiHandlerInfo.signCheckHandler != nil {
		if !apiHandlerInfo.signCheckHandler.signCheck(c) {
			failHandler(c, http.StatusBadRequest, ErrInvalidSignature, "非法签名")
			return
		}
	}
	//处理数据解密
	if apiHandlerInfo.cryptoHandler != nil && apiHandlerInfo.cryptoHandler.decryptImpl != nil {
		if !apiHandlerInfo.cryptoHandler.decryptImpl(c) {
			failHandler(c, http.StatusBadRequest, ErrCryptoError, "数据加密错误")
			return
		}
	}
	// 处理 CORS
	if apiHandlerInfo.corsHandler != nil &&
		apiHandlerInfo.corsHandler.CheckOriginByRequest(c.Request) {
		apiHandlerInfo.corsHandler.WriteCORSHeader(c)
	}

	// 处理 OPTIONS 请求
	if c.Request.Method == http.MethodOptions {
		c.PureJSON(200, gin.H{})
		return
	}

	reqId, level, er := getReqInfo(c)
	util.CheckError(er)

	apiContext := new(ApiContext)
	apiContext.ClientIP = c.ClientIP()
	apiContext.Request = c.Request
	apiContext.respHeaders = make(map[string]string, 4)

	// 请求追踪
	apiContext.ReqId = reqId
	apiContext.ReqLevel = level + 1

	// 验证请求是否在有效时间内
	// prehandler之前设置
	apiContext.APIConfig.Expires = apiHandlerInfo.expire

	defer func() {
		s := time.Since(start).Seconds()
		requestLatency.WithLabelValues(apiKey).Observe(s)
	}()
	// 处理请求
	request := reflect.New(apiHandlerInfo.reqType).Interface()

	var err error
	if apiHandlerInfo.pt == formType {
		err = c.ShouldBindQuery(request)
	} else {
		err = c.ShouldBindBodyWith(request, binding.JSON)
	}

	if err != nil {
		failHandler(c, http.StatusBadRequest, ErrInvalidParam, err.Error())
		return
	}

	if len(apiHandlerInfo.preHandlers) != 0 {
		apiContext.Keys = make(map[string]interface{}, 4)

		for _, handler := range apiHandlerInfo.preHandlers {
			resp := handler(c, apiContext)
			if resp != nil && resp.Code != 0 {
				failHandler(c, http.StatusOK, resp.Code, resp.Message)
				return
			}
		}
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
			// 正常响应,返回数据加密
			if apiHandlerInfo.cryptoHandler != nil && apiHandlerInfo.cryptoHandler.encryptImpl != nil {
				response = apiHandlerInfo.cryptoHandler.encryptImpl(c, response)
			}
		} else {
			err = ret[1].Interface().(error)
			response = Resp{Code: ErrLogicError, Message: err.Error()}
		}

		// access 日志处理
		log.GetLogger("access").Info("api", zap.String("reqId", reqId), zap.String("api", apiKey), zap.Int("useTime(ms)", int(time.Since(start).Milliseconds())), zap.Any("request", request), zap.Any("context", apiContext), zap.Any("Response", response))

		// 写入 Header
		for key, val := range apiContext.respHeaders {
			c.Header(key, val)
		}

		switch apiHandlerInfo.respType {
		case XmlType:
			c.XML(200, response)
		case StringType:
			c.String(200, "%v", response)
		case TextHtmlType:
			respData, ok := response.([]byte)
			if !ok {
				panic("content-type is text/html,resp type must be []byte")
			}
			c.Data(200, "text/html; charset=utf-8", respData)
		case OctetStreamType:
			respData, ok := response.([]byte)
			if !ok {
				panic("content-type is application/octet-stream,resp type must be []byte")
			}
			c.Data(200, "application/octet-stream; charset=utf-8", respData)
		case JsonStreamType:
			respData, ok := response.([]byte)
			if !ok {
				panic("content-type is application/json,resp type must be []byte")
			}
			c.Data(200, "application/json; charset=utf-8", respData)
		default:
			c.PureJSON(200, response) // 默认是json，但是如果返回的是二进制的json串，此方法会编码成base64格式

		}
	}
}

func failHandler(c *gin.Context, status int, code int, message string) {

	ua := c.Request.UserAgent()

	if 0 == strings.Index(ua, "curl") {
		sb := strings.Builder{}
		sb.WriteString("为了方便调试：curl 请求异常时返回内容不是json， 而是纯文本\n")
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("Code: %d\n", code))
		sb.WriteString("\n")
		sb.WriteString(message)
		sb.WriteString("\n")

		c.Data(status, "text/plain charset=utf-8", []byte(sb.String()))
		c.Abort()
	} else {
		c.AbortWithStatusJSON(status, Resp{Code: code, Message: message})

	}
	invalidRequestCounter.WithLabelValues(c.Request.URL.Path, strconv.Itoa(code)).Inc()

	body, _ := c.Get(gin.BodyBytesKey)
	log.GetLogger("access_error").Info(c.Request.URL.Path,
		zap.Int("code", code),
		zap.String("message", message),
		zap.String("url", c.Request.URL.String()),
		zap.String("body", cast.ToString(body)),
		zap.String("referer", c.Request.Referer()),
		zap.String("reqId", cast.ToString(c.Keys[perf.ReqIdKey])),
	)
}

func getAPIKey(path string) string {
	api := strings.ReplaceAll(path[5:], "/", ".")
	return api
}

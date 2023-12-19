package gateway

import (
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// ErrOK 正确
	ErrOK = 0

	// ErrInvalidParam 请求参数错误
	ErrInvalidParam = 1

	// ErrExpiredRequest 请求已经过期
	ErrExpiredRequest = 2

	// ErrInvalidAppKey 非法的appkey
	ErrInvalidAppKey = 3

	// ErrInvalidSignature 错误的签名
	ErrInvalidSignature = 4

	// ErrAppNotAuthed App 未授权调用该接口
	ErrAppNotAuthed = 5

	// ErrLogicError
	ErrLogicError = 6

	// ErrUnknowError 未知服务错误
	ErrUnknowError = 7

	// ErrCryptoError 数据加密异常
	ErrCryptoError = 8
)

type Resp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ApiContext struct {
	AppKey    string `json:"appKey"`
	Signature string `json:"signature"`
	Nonce     string `json:"nonce"`
	DeviceID  string `json:"deviceId"`
	AppID     string `json:"appId"`
	Host      string `json:"-"`
	UserAgent string `json:"-"`
	Partner   string `json:"partner"`
	ClientIP  string `json:"-"`

	Timestamp int64         `json:"timestamp"`
	Request   *http.Request `json:"-"`

	// For trace req
	ReqId    string `json:"-"`
	ReqLevel int    `json:"-"`

	// Keys is a key/value pair exclusively for the context of each request.
	Keys      map[string]interface{} `json:"-"`
	APIConfig struct {
		Expires time.Duration
	} `json:"-"`

	// response headers
	respHeaders map[string]string `json:"-"`
	m           sync.Mutex
}

func (c *ApiContext) WriteHeader(key, val string) {
	if c.respHeaders != nil {
		c.respHeaders[key] = val
	}
}

type Handler interface {
	Handler(*ApiContext) (interface{}, error)
}

type Pair struct {
	First  string
	Second string
}

type PreHandler func(*gin.Context, *ApiContext) *Resp

type paramType int

const (
	jsonType paramType = 0
	formType paramType = 1
)

type respType int

const (
	JsonType respType = iota
	XmlType
	StringType
	TextHtmlType
	OctetStreamType
	JsonStreamType
)

type HandlerInfo struct {
	reqType reflect.Type
	// reqParaMap map[string]Pair
	handler reflect.Type
	// 接口签名验证时间的有效时间长度
	expire time.Duration

	pt       paramType
	respType respType
	// 加解密处理
	cryptoHandler *CryptoHandler
	//签名校验处理
	signCheckHandler   *SignCheckHandler
	signCheckHandlerV2 *SignCheckHandlerV2
	// preHandlers 在业务Handler 处理前，定义的预处理
	preHandlers []PreHandler

	// CORS 处理器
	corsHandler *CORSHandler

	// extInfo 扩展属性
	extInfo map[string]string
}

type FieldInfo struct {
	name     string
	desc     string
	typ      string
	required bool
	note     string
}

// API 接口信息
type API struct {
	LineNum  int
	Key      string
	Name     string
	Summary  string
	Group    string
	ReqType  reflect.Type
	Request  *DTOInfo
	Response *DTOInfo
}

type TypeInfo struct {
	name   string
	fields []*FieldInfo
}

type DTOInfo struct {
	fields []*FieldInfo
	types  []*TypeInfo
}

// ============== 以下方法实现golang的标准context接口 ============

// Deadline  ...
func (*ApiContext) Deadline() (deadline time.Time, ok bool) {
	return
}

// Done 。。。
func (*ApiContext) Done() <-chan struct{} {
	return nil
}

// Err ...
func (*ApiContext) Err() error {
	return nil
}

// Value ...
func (*ApiContext) Value(key interface{}) interface{} {
	return nil
}

// =======================================================

// GetReqId ...
func (c *ApiContext) GetReqId() string {

	return c.ReqId
}

// GetReqLevel ...
func (c *ApiContext) GetReqLevel() int {

	return c.ReqLevel
}

// GetClientIP ...
func (c *ApiContext) GetClientIP() string {

	return c.ClientIP
}

// SetKV ...
func (c *ApiContext) SetKV(key string, value interface{}) {
	c.m.Lock()
	defer c.m.Unlock()
	c.Keys[key] = value
}

// GetKV ...
func (c *ApiContext) GetKV(key string) interface{} {
	c.m.Lock()
	defer c.m.Unlock()
	return c.Keys[key]
}

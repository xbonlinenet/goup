package gateway

import (
	"net/http"
	"reflect"
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
)

type Resp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ApiContext struct {
	AppKey    string `json:"appKey"`
	Signature string `json:"signature"`
	Timestamp int64  `json:"timestamp"`
	Nonce     string `json:"nonce"`
	DeviceID  string `json:"deviceId"`
	AppID     string `json:"appId"`
	Partner   string `json:"partner"`

	ClientIP string        `json:"-"`
	Request  *http.Request `json:"-"`

	// Keys is a key/value pair exclusively for the context of each request.
	Keys map[string]interface{} `json:"-"`
}

type Handler interface {
	Handler(*ApiContext) (interface{}, error)
}

type Pair struct {
	First  string
	Second string
}

type PreHandler func(*gin.Context, *ApiContext) *Resp

type HandlerInfo struct {
	reqType reflect.Type
	// reqParaMap map[string]Pair
	handler reflect.Type
	// 接口签名验证时间的有效时间长度
	expire time.Duration

	// preHandler 在业务Handler 处理前，定义的预处理
	preHandler PreHandler
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

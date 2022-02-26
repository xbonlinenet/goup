package gateway

import (
	"time"
)

// An Option configures
type Option interface {
	apply(*HandlerInfo)
}

// optionFunc wraps a func so it satisfies the Option interface.
type optionFunc func(*HandlerInfo)

func (f optionFunc) apply(log *HandlerInfo) {
	f(log)
}

// Expired 接口请求过期时间，过滤异常的客户端
func Expired(duration time.Duration) Option {
	return optionFunc(func(handler *HandlerInfo) {
		handler.expire = duration
	})
}

// HandlerFunc 设置 PreHandler，可用于统一登录鉴权使用
func HandlerFunc(handlerFunc PreHandler) Option {
	return optionFunc(func(handler *HandlerInfo) {
		handler.preHandlers = append(handler.preHandlers, handlerFunc)
	})
}

func FormParam() Option {
	return optionFunc(func(handler *HandlerInfo) {
		handler.pt = formType
	})
}

// ResponseXML 返回 xml 类型内容
func ResponseXML() Option {
	return optionFunc(func(handler *HandlerInfo) {
		handler.respType = XmlType
	})
}
func ResponseString() Option {
	return optionFunc(func(handler *HandlerInfo) {
		handler.respType = StringType
	})
}

// RespContentType 指定响应类型，传入定义的常量
func RespContentType(contentType respType) Option {
	return optionFunc(func(handler *HandlerInfo) {
		handler.respType = contentType
	})
}

// ExtInfo 设置额外的说明信息
func ExtInfo(info map[string]string) Option {
	return optionFunc(func(handler *HandlerInfo) {
		handler.extInfo = info
	})
}

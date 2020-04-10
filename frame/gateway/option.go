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
		handler.preHandler = handlerFunc
	})
}

func FormParam() Option {
	return optionFunc(func(handler *HandlerInfo) {
		handler.pt = formType
	})
}

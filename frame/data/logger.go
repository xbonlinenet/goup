package data

import (
	"github.com/xbonlinenet/goup/frame/log"
)

// NewWriter writer 构造函数
func NewWriter() writer {
	return writer{}
}

type writer struct {
}

// Println 格式化打印日志
func (w writer) Print(v ...interface{}) {
	log.Default().Sugar().Info(v)
}

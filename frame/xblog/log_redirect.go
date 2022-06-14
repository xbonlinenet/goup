package xblog

import (
	"github.com/xbonlinenet/goup/frame/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Just add a different package name, so to avoid name conflict from std package 'log'
// all the exported function will call relative function in pkg log

func Init() error {
	return log.Init()
}

func GetLogger(name string) *zap.Logger {
	return log.GetLogger(name)
}

func Default() *zap.Logger {
	return log.Default()
}

func GetLogFields(keyValuePairs ...interface{}) []zap.Field{
	return log.GetLogFields(keyValuePairs...)
}

func Sugar() *zap.SugaredLogger {
	return log.Sugar()
}

func With(fields ...zapcore.Field) *zap.SugaredLogger{
	return log.With(fields...)
}
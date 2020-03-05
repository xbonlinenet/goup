package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func sugar() *zap.SugaredLogger {
	return Default().Sugar()
}

func Sugar() *zap.SugaredLogger {
	return Default().Sugar()
}

// // Debugf uses fmt.Sprintf to log a templated message.
// func Debugf(template string, args ...interface{}) {
// 	sugar().Debugf(template, args...)
// }

// // Infof uses fmt.Sprintf to log a templated message.
// func Infof(template string, args ...interface{}) {
// 	sugar().Infof(template, args...)
// }

// // Info uses fmt.Sprint to construct and log a message.
// func Info(args ...interface{}) {
// 	sugar().Info(args...)
// }

// // Warnf uses fmt.Sprintf to log a templated message.
// func Warnf(template string, args ...interface{}) {
// 	sugar().Warnf(template, args...)
// }

// // Errorf uses fmt.Sprintf to log a templated message.
// func Errorf(template string, args ...interface{}) {
// 	sugar().Errorf(template, args...)
// }

func With(fields ...zapcore.Field) *zap.SugaredLogger {
	return Default().With(fields...).Sugar()
}

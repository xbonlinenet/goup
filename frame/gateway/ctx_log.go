package gateway

import (
	"fmt"
	"github.com/xbonlinenet/goup/frame/log"
	"go.uber.org/zap"
)

// buildTrace 。。。
func (c *ApiContext) buildTrace() []zap.Field {
	trace := log.GetLogFields("reqId", c.ReqId, "reqLevel", c.ReqLevel)
	return trace
}

// Error ...
func (c *ApiContext) Error(msg string, fields ...zap.Field) {
	log.Default().With(c.buildTrace()...).Error(msg, fields...)
}

// Errorf ...
func (c *ApiContext) Errorf(format string, msg string, fields ...zap.Field) {
	c.Error(fmt.Sprintf(format, msg), fields...)
}

// Info ...
func (c *ApiContext) Info(msg string, fields ...zap.Field) {
	log.Default().With(c.buildTrace()...).Info(msg, fields...)
}

// Infof ...
func (c *ApiContext) Infof(format string, msg string, fields ...zap.Field) {
	c.Info(fmt.Sprintf(format, msg), fields...)
}

// Debug ...
func (c *ApiContext) Debug(msg string, fields ...zap.Field) {
	log.Default().With(c.buildTrace()...).Debug(msg, fields...)
}

// Debugf ...
func (c *ApiContext) Debugf(format string, msg string, fields ...zap.Field) {
	c.Debug(fmt.Sprintf(format, msg), fields...)
}

// Warn ...
func (c *ApiContext) Warn(msg string, fields ...zap.Field) {
	log.Default().With(c.buildTrace()...).Warn(msg, fields...)
}

// Warnf ...
func (c *ApiContext) Warnf(format string, msg string, fields ...zap.Field) {
	c.Warn(fmt.Sprintf(format, msg), fields...)
}

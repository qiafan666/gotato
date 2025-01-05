package gface

import (
	"context"
	"fmt"
	"github.com/qiafan666/gotato/commons/gcommon"
	"go.uber.org/zap"
	"log"
)

// Logger 定义一个通用的业务日志接口
// 已接入的gotato组件有，gapp，timer，module，gpromise，gtask
type Logger interface {
	ErrorF(ctx context.Context, format string, args ...interface{})
	WarnF(ctx context.Context, format string, args ...interface{})
	DebugF(ctx context.Context, format string, args ...interface{})
	InfoF(ctx context.Context, format string, args ...interface{})
	Logger() *zap.SugaredLogger
	Prefix() string
}

// ------------------------ example ------------------------

// LoggerImpl 实现了Logger接口,也可在项目中自己实现Logger接口
type LoggerImpl struct {
	module string
	logger *zap.SugaredLogger
}

// NewLogger 创建一个新的Logger实例，接受一个module名称作为前缀
func NewLogger(module string, logger *zap.SugaredLogger) *LoggerImpl {
	return &LoggerImpl{
		module: module,
		logger: logger,
	}
}
func (l *LoggerImpl) ErrorF(ctx context.Context, format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[ERROR] [%s] ", l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	} else {
		l.Logger().Errorf(fmt.Sprintf(l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	}
}
func (l *LoggerImpl) WarnF(ctx context.Context, format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[WARN] [%s] ", l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	} else {
		l.Logger().Warnf(fmt.Sprintf(l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	}
}
func (l *LoggerImpl) InfoF(ctx context.Context, format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[INFO] [%s] ", l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	} else {
		l.Logger().Infof(fmt.Sprintf(l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	}
}
func (l *LoggerImpl) DebugF(ctx context.Context, format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[DEBUG] [%s] ", l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	} else {
		l.Logger().Debugf(fmt.Sprintf(l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	}
}
func (l *LoggerImpl) Logger() *zap.SugaredLogger {
	return l.logger
}
func (l *LoggerImpl) Prefix() string {
	return fmt.Sprintf("[module:%s] ", l.module)
}

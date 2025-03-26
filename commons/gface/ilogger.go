package gface

import (
	"context"
	"fmt"
	"github.com/qiafan666/gotato/commons/gcommon"
	"go.uber.org/zap"
	"log"
)

// ILogger 定义一个通用的业务日志接口
// 已接入的gotato组件有，gapp，timer，module，gpromise，gtask
type ILogger interface {
	ErrorF(ctx context.Context, format string, args ...interface{})
	WarnF(ctx context.Context, format string, args ...interface{})
	DebugF(ctx context.Context, format string, args ...interface{})
	InfoF(ctx context.Context, format string, args ...interface{})
	Logger() *zap.SugaredLogger
	Prefix() string
}

// ------------------------ example ------------------------

// loggerImpl 实现了Logger接口,也可在项目中自己实现Logger接口
type loggerImpl struct {
	module string
	logger *zap.SugaredLogger
}

// NewLogger 创建一个新的Logger实例，接受一个module名称作为前缀
func NewLogger(module string, logger *zap.SugaredLogger) ILogger {
	return &loggerImpl{
		module: module,
		logger: logger,
	}
}
func (l *loggerImpl) ErrorF(ctx context.Context, format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[ERROR] %s ", l.Prefix())+gcommon.GetRequestIdFormat(ctx)+format, args...)
	} else {
		l.Logger().Errorf(l.Prefix()+gcommon.GetRequestIdFormat(ctx)+format, args...)
	}
}
func (l *loggerImpl) WarnF(ctx context.Context, format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[WARN] %s ", l.Prefix())+gcommon.GetRequestIdFormat(ctx)+format, args...)
	} else {
		l.Logger().Warnf(l.Prefix()+gcommon.GetRequestIdFormat(ctx)+format, args...)
	}
}
func (l *loggerImpl) InfoF(ctx context.Context, format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[INFO] %s ", l.Prefix())+gcommon.GetRequestIdFormat(ctx)+format, args...)
	} else {
		l.Logger().Infof(l.Prefix()+gcommon.GetRequestIdFormat(ctx)+format, args...)
	}
}
func (l *loggerImpl) DebugF(ctx context.Context, format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[DEBUG] %s ", l.Prefix())+gcommon.GetRequestIdFormat(ctx)+format, args...)
	} else {
		l.Logger().Debugf(l.Prefix()+gcommon.GetRequestIdFormat(ctx)+format, args...)
	}
}
func (l *loggerImpl) Logger() *zap.SugaredLogger {
	return l.logger
}
func (l *loggerImpl) Prefix() string {
	return fmt.Sprintf("[module:%s] ", l.module)
}

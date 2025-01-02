package gface

import (
	"fmt"
	"go.uber.org/zap"
	"log"
)

// Logger 定义一个通用的业务日志接口
// 已接入的gotato组件有，gapp，timer，module，gpromise，gtask
type Logger interface {
	ErrorF(format string, args ...interface{})
	WarnF(format string, args ...interface{})
	DebugF(format string, args ...interface{})
	InfoF(format string, args ...interface{})
	Logger() *zap.SugaredLogger
	Prefix() string
}

// ------------------------ example ------------------------
type logger struct{}

func (l *logger) ErrorF(format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[ERROR] [%s] ", l.Prefix())+format, args...)
	} else {
		l.Logger().Errorf(fmt.Sprintf(l.Prefix())+format, args...)
	}
}
func (l *logger) WarnF(format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[WARN] [%s] ", l.Prefix())+format, args...)
	} else {
		l.Logger().Warnf(fmt.Sprintf(l.Prefix())+format, args...)
	}
}
func (l *logger) InfoF(format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[INFO] [%s] ", l.Prefix())+format, args...)
	} else {
		l.Logger().Infof(fmt.Sprintf(l.Prefix())+format, args...)
	}
}
func (l *logger) DebugF(format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[DEBUG] [%s] ", l.Prefix())+format, args...)
	} else {
		l.Logger().Debugf(fmt.Sprintf(l.Prefix())+format, args...)
	}
}
func (l *logger) Logger() *zap.SugaredLogger {
	return nil
}
func (l *logger) Prefix() string {
	return "test"
}

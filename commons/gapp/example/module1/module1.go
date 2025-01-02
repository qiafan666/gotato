package module1

import (
	"fmt"
	"github.com/qiafan666/gotato/commons/gapp/chanrpc"
	"github.com/qiafan666/gotato/commons/gapp/example/def"
	"github.com/qiafan666/gotato/commons/gapp/module"
	"github.com/qiafan666/gotato/commons/gapp/timer/timermgr"
	"github.com/qiafan666/gotato/commons/gface"
	"go.uber.org/zap"
	"log"
)

var (
	GoLen       = 1000
	AsynCallLen = 10000
	ChanRPCLen  = 10000
)

type Module1 struct {
	skeleton module.ISkeleton
	timerMgr *timermgr.TimerMgr
}

func NewModule() *Module1 {
	return &Module1{
		skeleton: module.NewSkeleton(GoLen, ChanRPCLen, AsynCallLen, &logger{}),
	}
}

// OnInit 初始化
func (m *Module1) OnInit() error {
	m.initHandler()
	m.initTimer()
	return nil
}

// OnDestroy 销毁
func (m *Module1) OnDestroy() {
}

// Run 启动
func (m *Module1) Run(closeSig chan bool) {
	m.skeleton.Run(closeSig)
}

// Name 模块名字
func (m *Module1) Name() string {
	return def.TEST1
}

// ChanSrv 消息通道
func (m *Module1) ChanSrv() chanrpc.IServer {
	return m.skeleton.Server()
}

// Logger 日志
func (m *Module1) Logger() gface.Logger {
	return m.skeleton.Logger()
}

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
	return def.ZapLog
}
func (l *logger) Prefix() string {
	return "[module:module1] "
}

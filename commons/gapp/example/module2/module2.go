package module2

import (
	"fmt"
	"github.com/qiafan666/gotato/commons/gapp"
	"github.com/qiafan666/gotato/commons/gapp/chanrpc"
	"github.com/qiafan666/gotato/commons/gapp/example/def"
	"github.com/qiafan666/gotato/commons/gapp/module"
	"github.com/qiafan666/gotato/commons/gcommon/sval"
	"github.com/qiafan666/gotato/commons/gface"
	"go.uber.org/zap"
	"log"
	"time"
)

var (
	GoLen       = 10000
	AsynCallLen = 10000
	ChanRPCLen  = 10000
)

type Module2 struct {
	skeleton module.ISkeleton
}

// --------------------------------------模块初始化相关----------------------------------

func NewModule() *Module2 {
	return &Module2{
		skeleton: module.NewSkeleton(GoLen, ChanRPCLen, AsynCallLen, &logger{}),
	}
}

// OnInit 初始化
func (m *Module2) OnInit() error {
	m.ChanSrv().Register((*def.Test1Ntf)(nil), m.onHandleTest1)
	return nil
}

func (m *Module2) onHandleTest1(ci *chanrpc.ReqCtx) {
	//req := ci.Req.(*iproto.Test1Ntf)
	//fmt.Printf("onHandleTest1 msg:%+v\n", req)
}

// Run 启动
func (m *Module2) Run(closeSig chan bool) {
	m.skeleton.Run(closeSig)
}

// OnDestroy 销毁
func (m *Module2) OnDestroy() {
}

// Name 模块名字
func (m *Module2) Name() string {
	return def.TEST2
}

// ChanSrv 消息通道
func (m *Module2) ChanSrv() chanrpc.IServer {
	return m.skeleton.Server()
}

// Logger 日志
func (m *Module2) Logger() gface.Logger {
	return m.skeleton.Logger()
}

type logger struct{}

func (l *logger) ErrorF(format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[ERROR] [%s] ", l.Prefix())+format, args...)
	} else {
		l.Logger().Errorf(fmt.Sprintf("[%s] ", l.Prefix())+format, args...)
	}
}
func (l *logger) WarnF(format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[WARN] [%s] ", l.Prefix())+format, args...)
	} else {
		l.Logger().Warnf(fmt.Sprintf("[%s] ", l.Prefix())+format, args...)
	}
}
func (l *logger) InfoF(format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[INFO] [%s] ", l.Prefix())+format, args...)
	} else {
		l.Logger().Infof(fmt.Sprintf("[%s] ", l.Prefix())+format, args...)
	}
}
func (l *logger) DebugF(format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[DEBUG] [%s] ", l.Prefix())+format, args...)
	} else {
		l.Logger().Debugf(fmt.Sprintf("[%s] ", l.Prefix())+format, args...)
	}
}
func (l *logger) Logger() *zap.SugaredLogger {
	return def.ZapLog
}
func (l *logger) Prefix() string {
	return "module2"
}

// --------------------------------------模块初始化相关----------------------------------

// Cast 异步调用
func (m *Module2) Cast(modName string, msg any) {
	gapp.DefaultApp().GetChanSrv(modName).Cast(msg)
}

// AsyncCall 异步调用
func (m *Module2) AsyncCall(modName string, req any, callback chanrpc.Callback, ctx sval.M) {
	m.skeleton.Client().AsyncCall(gapp.DefaultApp().GetChanSrv(modName), req, callback, ctx)
}

// Call 同步调用，逻辑层面的Call都应该加上超时
func (m *Module2) Call(modName string, req any) *chanrpc.AckCtx {
	return m.skeleton.Client().CallT(gapp.DefaultApp().GetChanSrv(modName), req, 5*time.Second)
}

// CallActor 同步调用，逻辑层面的Call都应该加上超时
func (m *Module2) CallActor(modName string, actorID int64, req any) *chanrpc.AckCtx {
	return m.skeleton.Client().CallT(gapp.DefaultApp().GetActorChanSrv(modName, actorID), req, 5*time.Second)
}

// CastActor 同步调用，逻辑层面的Call都应该加上超时
func (m *Module2) CastActor(modName string, actorID int64, req any) {
	gapp.DefaultApp().GetActorChanSrv(modName, actorID).Cast(req)
}

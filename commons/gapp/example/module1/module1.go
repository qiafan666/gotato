package module1

import (
	"github.com/qiafan666/gotato/commons/gapp/chanrpc"
	"github.com/qiafan666/gotato/commons/gapp/example/def"
	"github.com/qiafan666/gotato/commons/gapp/module"
	"github.com/qiafan666/gotato/commons/gapp/timer/timermgr"
	"github.com/qiafan666/gotato/commons/gface"
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
		skeleton: module.NewSkeleton(GoLen, ChanRPCLen, AsynCallLen, gface.NewLogger(def.TEST1, def.ZapLog)),
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

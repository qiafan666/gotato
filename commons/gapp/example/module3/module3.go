package module3

import (
	"github.com/qiafan666/gotato/commons/gapp/actor"
	"github.com/qiafan666/gotato/commons/gapp/chanrpc"
	"github.com/qiafan666/gotato/commons/gapp/example/def"
	"github.com/qiafan666/gotato/commons/gapp/module"
	"github.com/qiafan666/gotato/commons/gface"
	"log"
	"time"
)

var (
	GoLen       = 10000
	AsynCallLen = 10000
	ChanRPCLen  = 10000
)

type Module3 struct {
	skeleton module.ISkeleton
	*actor.Mgr
}

// --------------------------------------模块初始化相关----------------------------------

func NewModule() *Module3 {
	creator := func(actorID int64) actor.IActor {
		return &testActor{
			id:       actorID,
			skeleton: module.NewSkeleton(10, 100000, 100000, gface.NewLogger("actor", def.ZapLog)),
		}
	}
	return &Module3{
		skeleton: module.NewSkeleton(GoLen, ChanRPCLen, AsynCallLen, gface.NewLogger(def.TEST3, def.ZapLog)),
		Mgr:      actor.NewMgr(creator, "", nil, 1),
	}
}

// OnInit 初始化
func (m *Module3) OnInit() error {
	m.StartActor(111, 1111, false)
	return nil
}

func (m *Module3) onHandleTest1(ci *chanrpc.ReqCtx) {
	req := ci.Req.(*def.Test1Ntf)
	log.Printf("onHandleTest1 msg:%+v", req)
}

// Run 启动
func (m *Module3) Run(closeSig chan bool) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case msg := <-m.skeleton.Server().ChanReq():
			m.skeleton.Server().Exec(msg)
		case ackCtx := <-m.skeleton.Client().ChanAck():
			m.skeleton.Client().Exec(ackCtx)
		case <-ticker.C:
			m.Logger().InfoF(nil, "module3 run")
		case <-closeSig:
			m.skeleton.Server().Close()
			m.skeleton.Client().Close()
			return
		}
	}
}

// OnDestroy 销毁
func (m *Module3) OnDestroy() {
	m.StopAllActor(true)
}

// Name 模块名字
func (m *Module3) Name() string {
	return def.TEST3
}

// ChanSrv 消息通道
func (m *Module3) ChanSrv() chanrpc.IServer {
	return m.skeleton.Server()
}

// Logger 日志
func (m *Module3) Logger() gface.ILogger {
	return m.skeleton.Logger()
}

package module3

import (
	"github.com/qiafan666/gotato/commons/gapp2/actor"
	"github.com/qiafan666/gotato/commons/gapp2/chanrpc"
	"github.com/qiafan666/gotato/commons/gapp2/example/def"
	"github.com/qiafan666/gotato/commons/gapp2/module"
	"log"
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
			skeleton: module.NewSkeleton(10, 100000, 100000),
		}
	}
	return &Module3{
		skeleton: module.NewSkeleton(GoLen, ChanRPCLen, AsynCallLen),
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
	m.skeleton.Run(closeSig)
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
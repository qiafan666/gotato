package module2

import (
	"github.com/qiafan666/gotato/commons/gapp2"
	"github.com/qiafan666/gotato/commons/gapp2/chanrpc"
	"github.com/qiafan666/gotato/commons/gapp2/example/def"
	"github.com/qiafan666/gotato/commons/gapp2/module"
	"github.com/qiafan666/gotato/commons/gcommon/sval"
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
		skeleton: module.NewSkeleton(GoLen, ChanRPCLen, AsynCallLen),
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

// --------------------------------------模块初始化相关----------------------------------

// Cast 异步调用
func (m *Module2) Cast(uid uint32, modName string, msg any) {
	gapp2.DefaultApp().GetChanSrv(modName).Cast(uid, msg)
}

// AsyncCall 异步调用
func (m *Module2) AsyncCall(uid uint32, modName string, req any, callback chanrpc.Callback, ctx sval.M) {
	m.skeleton.Client().AsyncCall(uid, gapp2.DefaultApp().GetChanSrv(modName), req, callback, ctx)
}

// Call 同步调用，逻辑层面的Call都应该加上超时
func (m *Module2) Call(uid uint32, modName string, req any) *chanrpc.AckCtx {
	return m.skeleton.Client().CallT(uid, gapp2.DefaultApp().GetChanSrv(modName), req, 5*time.Second)
}

// CallActor 同步调用，逻辑层面的Call都应该加上超时
func (m *Module2) CallActor(uid uint32, modName string, actorID int64, req any) *chanrpc.AckCtx {
	return m.skeleton.Client().CallT(uid, gapp2.DefaultApp().GetActorChanSrv(modName, actorID), req, 5*time.Second)
}

// CastActor 同步调用，逻辑层面的Call都应该加上超时
func (m *Module2) CastActor(uid uint32, modName string, actorID int64, req any) {
	gapp2.DefaultApp().GetActorChanSrv(modName, actorID).Cast(uid, req)
}

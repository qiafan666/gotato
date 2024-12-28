package module3

import (
	"github.com/qiafan666/gotato/commons/gapp2"
	"github.com/qiafan666/gotato/commons/gapp2/chanrpc"
	"github.com/qiafan666/gotato/commons/gapp2/example/def"
	"github.com/qiafan666/gotato/commons/gapp2/module"
	"log"
	"time"
)

type testActor struct {
	id       int64
	skeleton module.ISkeleton
}

func (ta *testActor) OnInit(initData any) error {
	log.Printf("actor%d OnInit with %v", ta.id, initData)
	ta.skeleton.Server().Register(&def.Test1ActorReq{}, ta.onTestMsg)
	return nil
}

func (ta *testActor) ChanSrv() chanrpc.IServer {
	return ta.skeleton.Server()
}

func (ta *testActor) Run(closeSig chan bool) {
	log.Printf("actor%d Run", ta.id)
	ta.skeleton.Run(closeSig)
}

func (ta *testActor) OnDestroy() {
	log.Printf("actor%d OnDestroy", ta.id)
}

func (ta *testActor) onTestMsg(reqCtx *chanrpc.ReqCtx) {
	// req := reqCtx.Req.(*iproto.Test1ActorReq)
	// fmt.Printf("actor%d receive testReq: %+v\n", ta.id, req)
	reqCtx.Reply(&def.Test1ActorAck{ErrCode: 2222})
}

func (ta *testActor) Call(uid uint32, modName string, req any) *chanrpc.AckCtx {
	return ta.skeleton.Client().CallT(uid, gapp2.DefaultApp().GetChanSrv(modName), req, 5*time.Second)
}

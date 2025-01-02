package module3

import (
	"context"
	"github.com/qiafan666/gotato/commons/gapp"
	"github.com/qiafan666/gotato/commons/gapp/chanrpc"
	"github.com/qiafan666/gotato/commons/gapp/example/def"
	"github.com/qiafan666/gotato/commons/gapp/module"
	"log"
	"time"
)

type testActor struct {
	id       int64
	skeleton module.ISkeleton
}

func (ta *testActor) OnInit(initData any) error {
	ta.skeleton.Logger().ErrorF(nil, "actor%d OnInit", ta.id)
	ta.skeleton.Server().Register(&def.Test1ActorReq{}, ta.onTestMsg)
	return nil
}

func (ta *testActor) ChanSrv() chanrpc.IServer {
	return ta.skeleton.Server()
}

func (ta *testActor) Run(closeSig chan bool) {
	ta.skeleton.Logger().ErrorF(nil, "actor%d Run", ta.id)
	ta.skeleton.Run(closeSig)
}

func (ta *testActor) OnDestroy() {
	log.Printf("actor%d OnDestroy", ta.id)
}

func (ta *testActor) onTestMsg(ctx context.Context, reqCtx *chanrpc.ReqCtx) {
	// req := reqCtx.Req.(*iproto.Test1ActorReq)
	// fmt.Printf("actor%d receive testReq: %+v\n", ta.id, req)
	reqCtx.Reply(ctx, &def.Test1ActorAck{ErrCode: 2222})
}

func (ta *testActor) Call(ctx context.Context, modName string, req any) *chanrpc.AckCtx {
	return ta.skeleton.Client().CallT(gapp.DefaultApp().GetChanSrv(modName), ctx, req, 5*time.Second)
}

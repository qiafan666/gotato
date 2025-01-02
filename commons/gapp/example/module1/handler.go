package module1

import (
	"context"
	"github.com/qiafan666/gotato/commons/gapp/chanrpc"
	"github.com/qiafan666/gotato/commons/gapp/example/def"
)

func (m *Module1) initHandler() {
	m.ChanSrv().Register((*def.Test1Ntf)(nil), m.onHandleTestNtf)
	m.ChanSrv().Register((*def.Test1Req)(nil), m.onHandleTestReq)
	m.ChanSrv().Register((*def.Test1CallReq)(nil), m.onHandleTestCallReq)
}

func (m *Module1) onHandleTestNtf(ctx context.Context, ci *chanrpc.ReqCtx) {
	req := ci.Req.(*def.Test1Ntf)
	m.Logger().DebugF(ctx, "onHandleTestNtf msg:%+v", req)
}

func (m *Module1) onHandleTestReq(ctx context.Context, ci *chanrpc.ReqCtx) {
	req := ci.Req.(*def.Test1Req)
	m.Logger().InfoF(ctx, "onHandleTestReq msg:%+v", req)
	var result int64
	for _, i := range req.T1 {
		result += i
	}
	ret := &def.Test1Ack{ErrCode: 222, Result: result}
	ci.Reply(ctx, ret)
}

func (m *Module1) onHandleTestCallReq(ctx context.Context, ci *chanrpc.ReqCtx) {
	// req := ci.Req.(*iproto.Test1CallReq)
	ret := &def.Test1CallAck{ErrCode: 333}
	m.Logger().WarnF(ctx, "onHandleTestCallReq msg:%+v", ret)
	ci.Reply(ctx, ret)
}

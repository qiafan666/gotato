package module1

import (
	"github.com/qiafan666/gotato/commons/gapp/chanrpc"
	"github.com/qiafan666/gotato/commons/gapp/example/def"
	"log"
)

func (m *Module1) initHandler() {
	m.ChanSrv().Register((*def.Test1Ntf)(nil), m.onHandleTestNtf)
	m.ChanSrv().Register((*def.Test1Req)(nil), m.onHandleTestReq)
	m.ChanSrv().Register((*def.Test1CallReq)(nil), m.onHandleTestCallReq)
}

func (m *Module1) onHandleTestNtf(ci *chanrpc.ReqCtx) {
	req := ci.Req.(*def.Test1Ntf)
	log.Printf("onHandleTest ntf msg:%+v", req)
}

func (m *Module1) onHandleTestReq(ci *chanrpc.ReqCtx) {
	req := ci.Req.(*def.Test1Req)
	log.Printf("onHandleTest req msg:%+v", req)
	ret := &def.Test1Ack{ErrCode: 222}
	ci.Reply(ret)
}

func (m *Module1) onHandleTestCallReq(ci *chanrpc.ReqCtx) {
	// req := ci.Req.(*iproto.Test1CallReq)
	ret := &def.Test1CallAck{ErrCode: 333}
	ci.Reply(ret)
}

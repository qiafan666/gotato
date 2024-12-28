package main

import (
	"fmt"
	"github.com/qiafan666/gotato/commons/gapp"
	"github.com/qiafan666/gotato/commons/gapp/chanrpc"
	"github.com/qiafan666/gotato/commons/gapp/example/def"
	"github.com/qiafan666/gotato/commons/gapp/example/module1"
	"github.com/qiafan666/gotato/commons/gapp/example/module2"
	"github.com/qiafan666/gotato/commons/gapp/example/module3"
	"github.com/qiafan666/gotato/commons/gapp/timer"
	"github.com/qiafan666/gotato/commons/gcommon/sval"
	"log"
	"time"
)

type StdLogger struct{}

func (l *StdLogger) ErrorF(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}
func (l *StdLogger) WarnF(format string, args ...interface{}) {
	log.Printf("[WARN] "+format, args...)
}
func (l *StdLogger) InfoF(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}
func (l *StdLogger) DebugF(format string, args ...interface{}) {
	log.Printf("[DEBUG] "+format, args...)
}

func main() {
	fmt.Println("test start")
	timer.Run(nil, &StdLogger{})
	m1 := module1.NewModule()
	m2 := module2.NewModule()
	m3 := module3.NewModule()
	gapp.DefaultApp().Start(&StdLogger{}, m1, m2, m3)

	// m1.ChanSrv().Cast(&iproto.Test1Ntf{PlayerID: 111, Name: "ning1", T1: []int64{1, 2, 3}})

	// 异步消息
	m2.Cast(def.TEST1, &def.Test1Ntf{PlayerID: 111, Name: "ning1", T1: []int64{1, 2, 3}})

	// 异步回调
	m2.AsyncCall(def.TEST1, &def.Test1Req{
		PlayerID: 222,
		Name:     "ning2",
		T1:       []int64{2, 3, 4},
	}, func(ackCtx *chanrpc.AckCtx) {
		if ackCtx.Err != nil {
			return
		}
		ack := ackCtx.Ack.(*def.Test1Ack)
		log.Printf("async call:%+v", ack)
	}, nil)

	// 异步回调带上下文
	m2.AsyncCall(def.TEST1, &def.Test1Req{
		PlayerID: 222,
		Name:     "ning2",
		T1:       []int64{3, 4, 5},
	}, func(ackCtx *chanrpc.AckCtx) {
		if ackCtx.Err != nil {
			return
		}
		ack := ackCtx.Ack.(*def.Test1Ack)
		log.Printf("async call with ctx:%+v %+v", ack, ackCtx.Ctx)
	}, sval.M{"111": sval.Int64(4444)})

	// 同步调用
	ret := m2.Call(def.TEST1, &def.Test1CallReq{PlayerID: 333, Name: "ning3", T1: []int64{3, 4, 5}}, 3234)
	if ret.Err != nil {
		log.Printf("call err:%v", ret.Err)
	} else {
		ack := ret.Ack.(*def.Test1CallAck)
		log.Printf("call ret:%+v", ack)
	}

	// 同步调用actor
	actorRet := m2.CallActor(def.TEST3, 111, &def.Test1ActorReq{PlayerID: 444, Name: "ning4", T1: []int64{4, 5, 6}})
	if actorRet.Err != nil {
		log.Printf("call actor err:%v", actorRet.Err)
	} else {
		ack := actorRet.Ack.(*def.Test1ActorAck)
		log.Printf("call actor ret:%+v", ack)
	}
	time.Sleep(3 * time.Second)
	timer.Stop()
	fmt.Println("test end")
}

package def

import "go.uber.org/zap"

var ZapLog *zap.SugaredLogger

const (
	TEST1 = "test1"
	TEST2 = "test2"
	TEST3 = "test3"
)

type Test1Ntf struct {
	PlayerID int64
	Name     string
	T1       []int64
}

type Test1Req struct {
	PlayerID int64
	Name     string
	T1       []int64
}

type Test1Ack struct {
	ErrCode int64
	Result  int64
}

type Test1CallReq struct {
	PlayerID int64
	Name     string
	T1       []int64
}

type Test1CallAck struct {
	ErrCode int64
	Data    []byte
}

type Test1ActorReq struct {
	PlayerID int64
	Name     string
	T1       []int64
}

type Test1ActorAck struct {
	ErrCode int64
}

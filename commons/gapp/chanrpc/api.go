package chanrpc

import (
	"errors"
	"github.com/qiafan666/gotato/commons/gcommon/sval"
	"time"
)

// chanrpc 对端错误定义，用于Cast/Call/AsyncCall，使用errors Wrap机制
// 因此请使用errros.Is(err, ErrPeerPanic)判断错误类型，而非直接"=="
var (
	ErrPeerChanRPCClosed    = errors.New("peer chanrpc closed")
	ErrPeerMsgNotRegsitered = errors.New("msg not register")
	ErrPeerChanRPCFull      = errors.New("peer chanrpc full")
	ErrPeerPanic            = errors.New("peer panic")
	ErrTimeout              = errors.New("timeout")
)

// IServer chanrpc server接口
type IServer interface {
	// API
	Register(msg any, f Handler)
	RegisterByName(msgName string, f Handler)
	Cast(req any, uid ...uint32)
	Call(req any, uid ...uint32) *AckCtx
	CallT(req any, timeout time.Duration, uid ...uint32) *AckCtx
	PendReq(reqCtx *ReqCtx, block bool)
	Len() int // 当前消息队列长度
}

// IClient chanrpc client接口
type IClient interface {
	// API
	Call(s IServer, req any, uid ...uint32) *AckCtx
	CallT(s IServer, req any, timeout time.Duration, uid ...uint32) *AckCtx
	AsyncCall(s IServer, req any, cb Callback, ctx sval.M, uid ...uint32)
	AsyncCallT(s IServer, req any, cb Callback, ctx sval.M, timeout time.Duration, uid ...uint32)
	ChanAck() chan *AckCtx
}

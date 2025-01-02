package chanrpc

import (
	"context"
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
	Cast(ctx context.Context, req any)
	Call(ctx context.Context, req any) *AckCtx
	CallT(ctx context.Context, req any, timeout time.Duration) *AckCtx
	PendReq(reqCtx *ReqCtx, block bool)
	Len() int // 当前消息队列长度
}

// IClient chanrpc client接口
type IClient interface {
	// API
	Call(s IServer, ctx context.Context, req any) *AckCtx
	CallT(s IServer, ctx context.Context, req any, timeout time.Duration) *AckCtx
	AsyncCall(s IServer, ctx context.Context, req any, cb Callback, m sval.M)
	AsyncCallT(s IServer, ctx context.Context, req any, cb Callback, m sval.M, timeout time.Duration)
	ChanAck() chan *AckCtx
}

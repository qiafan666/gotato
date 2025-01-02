package chanrpc

import (
	"context"
	"errors"
	"github.com/qiafan666/gotato/commons/gapp/logger"
	"github.com/qiafan666/gotato/commons/gapp/timer"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/gcommon/sval"
	"github.com/qiafan666/gotato/commons/gid"
	"time"
)

const (
	defaultTimeout = 3 * time.Second
)

type pendingAsyncCall struct {
	reqCtx     *ReqCtx
	cb         Callback
	m          sval.M
	deadlineTs int64
	expired    bool // 是否已经检查过超时
	ctx        context.Context
}

// Client 客户端
type Client struct {
	chanSyncAck       chan *AckCtx // 同步调用结果通道
	chanAsyncAck      chan *AckCtx // 异步调用结果通道
	closeSig          chan bool
	pendingAsyncCalls map[int64]*pendingAsyncCall
}

// NewClient 新建客户端 设置异步结果通道的最大缓冲值
func NewClient(l int) *Client {
	c := &Client{
		chanSyncAck:       make(chan *AckCtx, 1),
		chanAsyncAck:      make(chan *AckCtx, l),
		closeSig:          make(chan bool, 1),
		pendingAsyncCalls: make(map[int64]*pendingAsyncCall),
	}
	return c
}

// Call 同步调用
func (c *Client) Call(s IServer, ctx context.Context, req any) *AckCtx {
	return c.CallT(s, ctx, req, defaultTimeout)
}

// CallT 同步带超时调用
func (c *Client) CallT(s IServer, ctx context.Context, req any, timeout time.Duration) *AckCtx {
	reqID := gid.ID()
	reqCtx := &ReqCtx{
		reqID:   reqID,
		id:      MsgID(req),
		Req:     req,
		chanAck: c.chanSyncAck,
		ctx:     ctx,
	}
	timer.NewSysDelegate().NewTimer(0, reqID, time.Now().UnixMilli()+timeout.Milliseconds(), func(i int64) {
		logger.DefaultLogger.WarnF(ctx, "chanrpc Client CallT timeout req:%+v msg id:%v server msg len:%v stat name:%s", req, reqCtx.id, s.Len(), reqCtx.GetStatName())
		reqCtx.PendAck(&AckCtx{
			reqID: reqID,
			Err:   ErrTimeout,
			ctx:   ctx,
		})
	})
	s.PendReq(reqCtx, true)
	ackCtx := <-c.chanSyncAck
	if ackCtx.reqID != reqID { // 在debug模式下会出现，超时Ack和本身的Ack都返回的情况，应该直接丢掉
		logger.DefaultLogger.WarnF(ctx, "chanrpc Client CallT rpc client call error")
		ackCtx = <-c.chanSyncAck
	}
	timer.NewSysDelegate().CancelTimer(reqID)
	return ackCtx
}

// AsyncCall 异步调用，使用默认超时，函数本身不返回error，所有的error都在回调中处理
func (c *Client) AsyncCall(s IServer, ctx context.Context, req any, cb Callback, m sval.M) {
	c.AsyncCallT(s, ctx, req, cb, m, defaultTimeout)
}

// AsyncCallT 异步调用，函数本身不返回error，所有的error都在回调中处理
func (c *Client) AsyncCallT(s IServer, ctx context.Context, req any, cb Callback, m sval.M, timeout time.Duration) {
	if c.chanAsyncAck == nil || cap(c.chanAsyncAck) == 0 {
		ackCtx := &AckCtx{
			Err: errors.New("invalid asyncCallLen"),
		}
		cb(ackCtx)
		return
	}
	reqID := gid.ID()
	reqCtx := &ReqCtx{
		reqID:   reqID,
		id:      MsgID(req),
		Req:     req,
		chanAck: c.chanAsyncAck,
		ctx:     ctx,
	}
	// 复用唯一请求ID，作为TimerID
	deadlineTs := time.Now().UnixMilli() + timeout.Milliseconds()
	timer.NewSysDelegate().NewTimer(0, reqID, deadlineTs, func(_ int64) {
		reqCtx.PendAck(&AckCtx{
			reqID: reqID,
			Err:   ErrTimeout,
			ctx:   ctx,
		})
	})
	c.pendingAsyncCalls[reqID] = &pendingAsyncCall{
		reqCtx:     reqCtx,
		cb:         cb,
		m:          m,
		deadlineTs: deadlineTs,
	}
	s.PendReq(reqCtx, false)
}

// 清理严重超时的异步请求，主要做兜底处理，防止内存泄露
func (c *Client) checkExpiredAsyncReqs() {
	nowTs := time.Now().UnixMilli()
	for reqID, info := range c.pendingAsyncCalls {
		if nowTs > info.deadlineTs {
			if info.expired {
				continue
			}
			logger.DefaultLogger.ErrorF(info.ctx, "chanrpc Client expired asyncReq: reqID: %d, reqName: %v", reqID, info.reqCtx.GetStatName())
			pendOk := info.reqCtx.PendAck(&AckCtx{
				Err:   ErrTimeout,
				reqID: reqID,
				ctx:   info.ctx,
			})
			// 如果 pend 成功，则标记下，否则下一轮还会重复pend
			if pendOk {
				info.expired = true
			}
			// 如果 pend 失败，则本轮啥也不做，交给下一轮检查
		}
	}
}

// Exec 执行回调
func (c *Client) Exec(ackCtx *AckCtx) {
	c.checkExpiredAsyncReqs()
	c.exec(ackCtx)
}

func (c *Client) exec(ackCtx *AckCtx) {
	req, ok := c.pendingAsyncCalls[ackCtx.reqID]
	if !ok {
		return
	}
	delete(c.pendingAsyncCalls, ackCtx.reqID)
	timer.NewSysDelegate().CancelTimer(ackCtx.reqID)
	ackCtx.M = req.m
	ackCtx.ctx = req.ctx
	func() {
		defer func() {
			if stack := gcommon.PrintPanicStack(); stack != "" {
				logger.DefaultLogger.ErrorF(req.ctx, "chanrpc Client exec panic error: %s", stack)
			}
		}()
		req.cb(ackCtx)
	}()
}

// Close 关闭client
func (c *Client) Close() {
	timeoutTimer := time.After(10 * time.Second)
	for len(c.pendingAsyncCalls) > 0 {
		select {
		case msg := <-c.chanAsyncAck:
			c.exec(msg)
		case <-timeoutTimer:
			logger.DefaultLogger.ErrorF(nil, "chanrpc Client Close timeout, discard pendAsyncCalls: %d", c.pendingAsyncCalls)
			for reqID, info := range c.pendingAsyncCalls {
				c.exec(&AckCtx{
					reqID: reqID,
					Err:   ErrTimeout,
					ctx:   info.ctx,
				})
			}
			return
		}
	}
}

// Idle 判断是否空闲
func (c *Client) Idle() bool {
	return len(c.pendingAsyncCalls) == 0
}

// ChanAck 返回异步结果Channel
func (c *Client) ChanAck() chan *AckCtx {
	return c.chanAsyncAck
}

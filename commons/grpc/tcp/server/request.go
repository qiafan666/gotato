package server

import (
	"context"
	"github.com/cloudwego/netpoll"
	"github.com/qiafan666/gotato/commons/gcache"
	"sync"
)

type Request struct {
	ctx       context.Context
	conn      netpoll.Connection
	closeOnce sync.Once
	closed    bool
	reqKeys   *gcache.ShardLockMap[string, bool]
}

func NewRequest(ctx context.Context, conn netpoll.Connection) *Request {
	req := &Request{
		ctx:     ctx,
		conn:    conn,
		reqKeys: gcache.NewShardLockMap[bool](),
	}
	return req
}

func (r *Request) NewKey(reqKey string) {
	r.reqKeys.Set(reqKey, true)
}

func (r *Request) RemoveKey(reqKey string) {
	r.reqKeys.Remove(reqKey)
}

func (r *Request) Keys() []string {
	return r.reqKeys.Keys()
}

func (r *Request) Context() context.Context {
	return r.ctx
}

func (r *Request) Conn() netpoll.Connection {
	return r.conn
}

func (r *Request) Close() {
	r.closeOnce.Do(func() {
		r.closed = true
		r.conn.Close()
	})
}

func (r *Request) IsClosed() bool {
	return r.closed
}

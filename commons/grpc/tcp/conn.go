// Package tcp 提供TCP连接的封装实现
package tcp

import (
	"context"
	"fmt"
	"github.com/qiafan666/gotato/commons/gcache"
	"github.com/qiafan666/gotato/commons/gerr"
	"github.com/qiafan666/gotato/commons/gface"
	"github.com/qiafan666/gotato/commons/gid"
	"github.com/qiafan666/gotato/commons/grpc"
	"github.com/qiafan666/gotato/commons/grpc/tcp/protocol"
	"github.com/qiafan666/gotato/commons/gticker"
	"net"
	"sync"
	"time"
)

// 定义默认的超时时间常量
const (
	defaultTimeout     = 2 * time.Second  // 默认超时时间
	defaultPingTimeout = 3 * time.Second  // 默认心跳超时时间,3s
	defaultIdleTimeout = 30 * time.Second // 默认空闲超时时间,30s
	defaultLiveTimeout = 2 * time.Hour    // 默认存活超时时间,2h
)

// RecvChan 接收响应的通道封装
type RecvChan struct {
	ChanId    string             // 通道ID
	Ch        chan *grpc.Message // 消息通道
	closeOnce sync.Once          // 确保只关闭一次
	closed    bool               // 关闭标志
}

// NewRecvChan 创建一个用于接收响应的channel
// chanId: 通道ID,需要确保唯一
func NewRecvChan(chanId string) *RecvChan {
	ch := make(chan *grpc.Message)
	return &RecvChan{
		ChanId: chanId,
		Ch:     ch,
	}
}

// Close 关闭接收通道
func (ch *RecvChan) Close() {
	ch.closeOnce.Do(func() {
		ch.closed = true
		close(ch.Ch)
	})
}

// Write 写入响应消息到通道
func (ch *RecvChan) Write(msg *grpc.Message) {
	if ch.closed {
		return
	}
	ch.Ch <- msg
}

// ConnOptions TCP连接配置选项
type ConnOptions struct {
	Timeout     time.Duration // 请求超时时间
	PingTimeout time.Duration // 心跳超时时间
	IdleTimeout time.Duration // 空闲超时时间
	LiveTimeout time.Duration // 最大存活时间
	Logger      gface.Logger  // 日志接口
}

// Conn TCP连接封装
type Conn struct {
	net.Conn // 内嵌net.Conn接口

	ctx    context.Context    // 上下文
	cancel context.CancelFunc // 取消函数

	connId          uint64    // 连接ID
	closed          bool      // 关闭标志
	closeOnce       sync.Once // 确保只关闭一次
	closeChanNotify chan any  // 关闭通知通道

	recvChans *gcache.ShardLockMap[string, *RecvChan] // 响应接收通道映射表

	protocol grpc.Protocol // RPC协议实现

	pingDeadline time.Time // 心跳超时截止时间
	idleDeadline time.Time // 空闲超时截止时间
	liveDeadline time.Time // 存活超时截止时间

	logger gface.Logger // 日志接口

	opt *ConnOptions // 连接配置选项
}

// NewConn 创建新的TCP连接
// ctx: 上下文
// conn: 原始TCP连接
// opt: 连接配置选项
func NewConn(ctx context.Context, conn net.Conn, opt *ConnOptions) *Conn {
	// 生成连接ID
	connId := gid.NewSerialId[uint64]().Id()

	// 设置跟踪ID
	ctx = context.WithValue(ctx, "connId", connId)
	ctx, cancel := context.WithCancel(ctx)

	c := &Conn{
		Conn:            conn,
		ctx:             ctx,
		cancel:          cancel,
		connId:          connId,
		closeOnce:       sync.Once{},
		closeChanNotify: make(chan any),
		recvChans:       gcache.NewShardLockMap[*RecvChan](),
		protocol:        protocol.New(),
		logger:          opt.Logger,
	}
	c.logger.DebugF(nil, "NewConn: connId=%+v, local=%s, remote=%s", connId, conn.LocalAddr().String(), conn.RemoteAddr().String())

	// 设置连接选项
	c.setOptions(opt)
	// 启动连接管理
	c.run(ctx)

	return c
}

// setOptions 设置连接选项,使用默认值填充未指定的选项
func (c *Conn) setOptions(opt *ConnOptions) {
	now := time.Now()
	if opt == nil {
		opt = new(ConnOptions)
	}
	if opt.Timeout == 0 {
		opt.Timeout = defaultTimeout
	}
	if opt.PingTimeout == 0 {
		opt.PingTimeout = defaultPingTimeout
	}
	if opt.IdleTimeout == 0 {
		opt.IdleTimeout = defaultIdleTimeout
	}
	if opt.LiveTimeout == 0 {
		opt.LiveTimeout = defaultLiveTimeout
	}

	c.opt = opt
	// 设置各类超时截止时间
	c.pingDeadline = now.Add(opt.PingTimeout).Add(1 * time.Second)
	c.idleDeadline = now.Add(opt.IdleTimeout)
	c.liveDeadline = now.Add(opt.LiveTimeout)
}

// run 启动连接管理
func (c *Conn) run(ctx context.Context) {
	// 启动消息接收协程
	go c.read(ctx)
	// 启动心跳检查协程
	go c.ping(ctx)
	// 启动空闲检查协程
	go c.idleCheck(ctx)
	// 启动存活时间检查协程
	go c.liveTimeoutCheck(ctx)

	// 监听上下文取消
	go func() {
		select {
		case <-ctx.Done():
			c.Close()
		}
	}()
}

// Close 关闭连接
func (c *Conn) Close() error {
	c.closeOnce.Do(func() {
		c.logger.DebugF(nil, "Conn.Close: start, connId=%+v", c.connId)
		c.closed = true
		c.cancel()
		c.Conn.Close()
		c.closeChanNotify <- c.connId
		close(c.closeChanNotify)
		// 延迟1秒关闭所有接收通道
		time.AfterFunc(time.Second, func() {
			if c.recvChans.Count() != 0 {
				c.recvChans.IterCb(func(key string, value interface{}) {
					v := value.(*RecvChan)
					v.Close()
				})
				c.recvChans.Clear()
			}
		})
	})
	return nil
}

// CloseNotify 返回关闭通知通道
func (c *Conn) CloseNotify() <-chan any {
	return c.closeChanNotify
}

// RemoveChan 移除接收通道
func (c *Conn) RemoveChan(ch *RecvChan) {
	c.recvChans.Remove(ch.ChanId)
}

// StoreChan 存储接收通道
func (c *Conn) StoreChan(ch *RecvChan) {
	c.recvChans.Set(ch.ChanId, ch)
}

// IsClosed 返回连接是否已关闭
func (c *Conn) IsClosed() bool {
	return c.closed
}

// Send 发送消息
// v: 要发送的消息
// ch: 用于接收响应的通道(可选)
func (c *Conn) Send(v *grpc.Message, ch *RecvChan) error {
	// 检查连接状态
	if c.ctx.Err() != nil || c.closed {
		return gerr.NewLang(gerr.UnKnowError)
	}

	// 编码消息
	encode, e := c.protocol.Encode(c.ctx, v)
	if e != nil {
		return gerr.NewLang(gerr.UnKnowError).WrapMsg(e.Error())
	}

	// 如果提供了接收通道,存储它
	if ch != nil {
		c.StoreChan(ch)
		c.logger.DebugF(nil, "Conn.Send: store recvChan, chanId=%+v, connId=%+v", ch.ChanId, c.connId)
	}

	// 发送消息
	_, err := c.Write(encode)
	if err != nil {
		c.logger.ErrorF(nil, "Conn.Send: write fail, connId=%+v, err=%+v", c.connId, err)
		c.Close()
		return gerr.NewLang(gerr.UnKnowError).WrapMsg(err.Error())
	}

	return nil
}

// read 消息接收循环
func (c *Conn) read(ctx context.Context) {
	defer c.Close()
	for {
		// 检查连接状态
		if ctx.Err() != nil || c.closed {
			return
		}

		// 解码接收到的消息
		v, e := c.protocol.Decode(ctx, c)
		if e != nil {
			continue
		}

		// 处理心跳消息
		if v.Command == grpc.CmdHeartbeat {
			go c.handleHeartbeat(v)
			continue
		}

		// 更新空闲连接deadline
		c.idleDeadline = time.Now().Add(c.opt.IdleTimeout)

		// 处理推送消息
		if v.PkgType == grpc.PkgTypePush {
			//TODO push
			c.logger.DebugF(nil, "Conn.read: receive push msg, msg=%s", v.Body)
			continue
		}

		// 查找并写入对应的接收通道
		chanId := fmt.Sprintf("%d", v.Sequence)
		c.recvChans.RemoveCb(chanId, func(key string, ch *RecvChan, exists bool) bool {
			if !exists {
				c.logger.DebugF(nil, "Conn.read: recvChan not exist, sequence=%+v", chanId)
				return true
			}
			if ch.closed {
				c.logger.DebugF(nil, "Conn.read: recvChan closed, sequence=%+v", chanId)
				return true
			}
			ch.Write(v)
			return true
		})
	}
}

// ping 心跳检查循环
func (c *Conn) ping(ctx context.Context) {
	//interval := time.Duration(c.pingTimeout) * time.Millisecond
	interval := 1500 * time.Millisecond
	pingTicker := gticker.NewTicker(interval, func() {
		// 检查心跳超时
		if c.pingDeadline.Before(time.Now()) {
			c.logger.DebugF(nil, "Conn.ping: ping deadline, connId=%+v", c.connId)
			c.Close()
			return
		}

		// 发送心跳请求
		req := c.pingRequest()
		c.Send(req, nil)
	})

	pingTicker.Run(ctx)
}

// idleCheck 检查连接空闲时间,关闭超时未使用的连接
func (c *Conn) idleCheck(ctx context.Context) {
	interval := 5 * time.Second
	idleCheckTicker := gticker.NewTicker(interval, func() {
		// 如果有未完成的请求,不检查空闲
		if c.recvChans.Count() != 0 {
			return
		}
		// 检查空闲超时
		if c.idleDeadline.Before(time.Now()) {
			c.logger.DebugF(nil, "Conn.idleCheck: idle deadline, connId=%+v", c.connId)
			c.Close()
			return
		}
	})

	idleCheckTicker.Run(ctx)
}

// liveTimeoutCheck 检查连接存活时长,关闭超时且无堆积任务的连接
func (c *Conn) liveTimeoutCheck(ctx context.Context) {
	interval := 1 * time.Minute
	liveTimeoutCheckTicker := gticker.NewTicker(interval, func() {
		// 如果有未完成的请求,不检查存活时间
		if c.recvChans.Count() != 0 {
			return
		}
		// 检查存活超时
		if c.liveDeadline.Before(time.Now()) {
			c.logger.DebugF(nil, "Conn.liveTimeoutCheck: live deadline, connId=%+v", c.connId)
			c.Close()
			return
		}
	})

	liveTimeoutCheckTicker.Run(ctx)
}

// pingRequest 创建心跳请求消息
func (c *Conn) pingRequest() *grpc.Message {
	req := &grpc.Message{
		Command:  grpc.CmdHeartbeat,
		PkgType:  grpc.PkgTypeRequest,
		ReqId:    0,
		Sequence: 0,
		Body:     nil,
		Heartbeat: &grpc.Heartbeat{
			Timeout: uint32(c.opt.PingTimeout.Milliseconds()),
		},
	}
	return req
}

// handleHeartbeat 处理心跳消息
func (c *Conn) handleHeartbeat(v *grpc.Message) {
	switch v.PkgType {
	case grpc.PkgTypeRequest:
		// 收到心跳请求,更新心跳截止时间并发送响应
		c.pingDeadline = time.Now().Add(time.Duration(v.Heartbeat.Timeout) * time.Millisecond).Add(1 * time.Second)
		sequence := grpc.NewSequence()
		reply := &grpc.Message{
			Command:   grpc.CmdHeartbeat,
			PkgType:   grpc.PkgTypeReply,
			ReqId:     uint64(sequence),
			Sequence:  sequence,
			Body:      nil,
			Heartbeat: nil,
		}
		err := c.Send(reply, nil)
		if err != nil {
			return
		}
	case grpc.PkgTypeReply:
		// 收到心跳响应,更新心跳截止时间
		c.pingDeadline = time.Now().Add(c.opt.PingTimeout).Add(1 * time.Second)
	default:
		// 其他包类型,忽略
	}
}

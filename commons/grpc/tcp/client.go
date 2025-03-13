// Package tcp 提供TCP协议的RPC客户端实现
package tcp

import (
	"context"
	"fmt"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/qiafan666/gotato/commons/gcast"
	"github.com/qiafan666/gotato/commons/gerr"
	"github.com/qiafan666/gotato/commons/gface"
	"github.com/qiafan666/gotato/commons/gpool"
	"github.com/qiafan666/gotato/commons/grpc"
	"github.com/qiafan666/gotato/commons/grpc/tcp/protocol"
	"net"
	"time"
)

// ClientOptions TCP客户端配置选项
type ClientOptions struct {
	MaxConn     int           // 最大连接数
	IdleConn    int           // 空闲连接数
	Timeout     time.Duration // 超时时间
	PingTimeout time.Duration // 心跳超时时间
	IdleTimeout time.Duration // 空闲超时时间
	LiveTimeout time.Duration // 读取超时时间
	RetryLimit  int           // 重试次数

	Hystrix HystrixOptions // hystrix配置
	Logger  gface.Logger   // 日志接口
}

// HystrixOptions hystrix配置
type HystrixOptions struct {
	Timeout                time.Duration // 超时时间
	SleepWindow            time.Duration // 重新尝试间隔
	MaxConcurrentRequests  int           // 最大并发
	RequestVolumeThreshold int           // 请求数量
	ErrorPercentThreshold  int           // 失败率
}

func (h *HystrixOptions) Parse() hystrix.CommandConfig {
	if h.Timeout.Milliseconds() == 0 {
		h.Timeout = 10 * time.Second
	}
	if h.SleepWindow.Milliseconds() == 0 {
		h.SleepWindow = 500 * time.Millisecond
	}
	if h.MaxConcurrentRequests <= 0 {
		h.MaxConcurrentRequests = 5000
	}
	if h.RequestVolumeThreshold <= 0 {
		h.RequestVolumeThreshold = 100
	}
	if h.ErrorPercentThreshold <= 0 {
		h.ErrorPercentThreshold = 50
	}

	conf := hystrix.CommandConfig{
		Timeout:                int(h.Timeout.Milliseconds()),
		SleepWindow:            int(h.SleepWindow.Milliseconds()),
		MaxConcurrentRequests:  h.MaxConcurrentRequests,
		RequestVolumeThreshold: h.RequestVolumeThreshold,
		ErrorPercentThreshold:  h.ErrorPercentThreshold,
	}

	return conf
}

// Client TCP客户端结构体
type Client struct {
	network string
	addr    string // 服务器地址

	opt *ClientOptions // 客户端配置选项

	protocol grpc.Protocol

	pool *gpool.Pool[*Conn] // 连接池

	hystrixCommandName string // hystrix命令名称

	logger gface.Logger
}

// NewClient 创建新的TCP客户端
// ctx: 上下文
// addr: 服务器地址
// opt: 客户端配置选项
func NewClient(ctx context.Context, addr string, opt *ClientOptions) *Client {
	c := &Client{
		network: "tcp",
		addr:    addr,
		opt:     opt,
		logger:  opt.Logger,
	}
	c.protocol = &protocol.TextRpcProtocol{}

	connPool, err := gpool.NewPool[*Conn](ctx, &gpool.Options[*Conn]{
		MaxSize:  uint(opt.MaxConn),
		InitSize: uint(opt.IdleConn),
		New: func() (*Conn, error) {
			conn, err := net.DialTimeout(c.network, c.addr, c.opt.Timeout)
			if err != nil {
				return nil, err
			}
			newConn := NewConn(ctx, conn, &ConnOptions{
				Timeout:     opt.Timeout,
				PingTimeout: opt.PingTimeout,
				IdleTimeout: opt.IdleTimeout,
				LiveTimeout: opt.LiveTimeout,
				Logger:      opt.Logger,
			})
			return newConn, nil
		},
	})
	if connPool == nil {
		c.logger.WarnF(nil, "Client.NewClient: init conn pool fail, err=%+v", err)
	}
	if err != nil {
		c.logger.WarnF(nil, "Client.NewClient: init conn pool fail, err=%+v", err)
	}
	c.pool = connPool

	c.hystrixCommandName = fmt.Sprintf("rpc_%s", addr)
	hystrix.ConfigureCommand(c.hystrixCommandName, opt.Hystrix.Parse())

	return c
}

// Do 执行RPC调用
// ctx: 上下文
// request: RPC请求消息
// 返回: RPC响应消息和错误信息
func (c *Client) Do(ctx context.Context, request *grpc.Message) (*grpc.Message, error) {
	var (
		res *grpc.Message
		err error
	)
	hystrix.Do(c.hystrixCommandName, func() error {
		res, err = c.do(ctx, request)
		if err != nil {
			c.logger.ErrorF(nil, "Client.Do: rpc call fail, reqId=%+v, err=%+v", request.ReqId, err)
			return gerr.WrapMsg(err, "rpc call fail")
		}
		return nil
	}, func(e error) error {
		c.logger.ErrorF(nil, "Client.Do: hystrix.Do fail, name=%+v, err=%+v", c.hystrixCommandName, e)
		res = nil
		err = gerr.WrapMsg(e, "hystrix do fail,this is a fallback error")
		return e
	})
	return res, err
}

// do 执行RPC调用的核心实现
// ctx: 请求上下文，用于超时控制和取消
// request: RPC请求消息
// 返回: RPC响应消息和错误信息
func (c *Client) do(ctx context.Context, request *grpc.Message) (*grpc.Message, error) {

	// 检查上下文是否已取消,避免无效调用
	if ctx.Err() != nil {
		return nil, gerr.WrapMsg(ctx.Err(), "context canceled")
	}

	// 从上下文获取重试次数,用于控制重试逻辑
	retry := gcast.ToInt(ctx.Value("retry"))

	// 从连接池获取可用连接
	// defer确保连接会被放回连接池
	conn, err := c.pool.Get(ctx)
	defer c.pool.Put(conn)

	// 处理获取连接失败的情况
	if err != nil {
		c.logger.ErrorF(nil, "Client.Do: get conn fail, reqId=%+v, err=%+v", request.ReqId, err)
		// 如果达到重试上限,关闭连接并返回错误
		if retry >= c.opt.RetryLimit {
			if conn != nil {
				conn.Close()
			}
			return nil, gerr.WrapMsg(err, "retry limit reached")
		}
		// 未达重试上限,递增重试计数并重试
		ctx = context.WithValue(ctx, "retry", retry+1)
		return c.do(ctx, request)
	}

	var (
		e  error
		ch *RecvChan = nil // 用于接收异步响应的通道
	)

	// 心跳包特殊处理:不需要等待响应
	if request.Command == grpc.CmdHeartbeat {
		e = conn.Send(request, nil)
	} else {
		// 非心跳包:创建接收通道并发送请求
		chanId := fmt.Sprintf("%d", request.Sequence)
		ch = NewRecvChan(chanId)
		defer ch.Close() // 确保通道会被关闭
		e = conn.Send(request, ch)
	}

	// 处理发送请求失败的情况
	if e != nil {
		c.logger.WarnF(nil, "Client.Do: send message fail, reqId=%+v, err=%+v, retry=%+v", request.ReqId, e, retry)
		// 发送失败时关闭连接,避免复用出问题的连接
		conn.Close()
		// 如果达到重试上限则返回错误
		if retry >= c.opt.RetryLimit {
			return nil, gerr.WrapMsg(e, "retry limit reached")
		}
		// 未达重试上限,递增重试计数并重试
		ctx = context.WithValue(ctx, "retry", retry+1)
		return c.do(ctx, request)
	}

	// 等待响应(仅非心跳包请求需要)
	if ch != nil {
		select {
		case <-ctx.Done():
			// 上下文取消,可能是父协程超时或取消
			c.logger.DebugF(nil, "Client.Do: context done, reqId=%+v", request.ReqId)
			return nil, gerr.New("context canceled")

		case resp := <-ch.Ch:
			// 收到响应
			if resp == nil {
				// 通道关闭,说明连接可能已断开
				c.logger.DebugF(nil, "Client.Do: receive chan closed, reqId=%+v", request.ReqId)
				return nil, gerr.New("receive chan closed")
			}
			return resp, nil

		case <-time.After(c.opt.Timeout):
			// 请求超时
			defer conn.RemoveChan(ch) // 清理超时的接收通道
			c.logger.DebugF(ctx, "Client.Do: receive timeout, reqId=%+v", request.ReqId)
			return nil, gerr.New("receive timeout")
		}
	}
	return nil, nil
}

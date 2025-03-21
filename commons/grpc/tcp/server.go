package tcp

import (
	"context"
	"fmt"
	"github.com/cloudwego/netpoll"
	"github.com/qiafan666/gotato/commons/gface"
	"github.com/qiafan666/gotato/commons/gid"
	"github.com/qiafan666/gotato/commons/grpc"
	"github.com/qiafan666/gotato/commons/grpc/tcp/protocol"
	"github.com/qiafan666/gotato/commons/grpc/tcp/server"
	"net"
	"syscall"
	"time"
)

type ServerOptions struct {
	Timeout time.Duration
	Logger  gface.Logger
}

type Server struct {
	addr string

	opt *ServerOptions

	protocol grpc.Protocol
	serialId *gid.SerialId[uint64]

	connManager *server.ConnManager

	handler grpc.Handler

	logger gface.Logger

	ch chan *grpc.Message
}

func NewServer(addr string, handler grpc.Handler, opt *ServerOptions) *Server {
	s := &Server{
		addr:    addr,
		handler: handler,
		opt:     opt,
	}
	s.protocol = protocol.New()
	s.serialId = gid.NewSerialId[uint64]()

	s.ch = make(chan *grpc.Message, 4096)
	s.connManager = server.NewConnManager(opt.Logger)

	s.logger = opt.Logger
	return s
}

func (s *Server) Run(ctx context.Context) {
	l := &net.ListenConfig{Control: func(network, address string, c syscall.RawConn) error {
		var opErr error
		err := c.Control(func(fd uintptr) {
			// syscall.SO_REUSEPORT ,在Linux下还可以指定端口重用
			opErr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
		})
		if err != nil {
			return err
		}
		return opErr
	}}

	ln, err := l.Listen(ctx, "tcp", s.addr)
	if err != nil {
		s.logger.ErrorF(nil, "Server.Run: listen tcp addr fail, err=%+v", err)
	}

	eventLoop, err := netpoll.NewEventLoop(
		s.handle,
		netpoll.WithOnPrepare(s.prepare),
		netpoll.WithOnConnect(s.connect),
		netpoll.WithReadTimeout(time.Second),
		netpoll.WithWriteTimeout(time.Second),
	)
	if err != nil {
		s.logger.ErrorF(nil, "Server.Run: NewEventLoop fail, err=%+v", err)
	}

	s.logger.InfoF(nil, "Server.Run: server start, addr=%s", s.addr)

	go func() {
		err = eventLoop.Serve(ln)
		if err != nil {
			s.logger.ErrorF(nil, "Server.Run: eventloop serve fail, err=%+v", err)
		}
	}()

	go func() {
		for {
			select {
			case msg := <-s.ch:
				// 收到响应消息
				reqKey := s.msgKey(msg)
				req, err := s.connManager.GetRequest(reqKey)
				if err != nil {
					s.logger.WarnF(nil, "Server.Run: GetRequest fail, err=%+v", err)
					continue
				}
				if req == nil || req.IsClosed() {
					s.logger.WarnF(nil, "Server.Run: request is nil or closed")
					continue
				}
				s.send(req.Context(), req.Conn(), msg)
			case <-ctx.Done():
				eventLoop.Shutdown(ctx)
				ln.Close()
				close(s.ch)
				s.logger.InfoF(nil, "Server.Run: server closed")
				return
			}

		}
	}()

	//go s.ping(ctx)
}

func (s *Server) handle(ctx context.Context, conn netpoll.Connection) error {
	msg, err := s.protocol.Decode(ctx, netpoll.NewIOReader(conn.Reader()))
	if err != nil {
		s.logger.InfoF(nil, "Server.handle: decode fail, err=%+v", err)
		s.connManager.CloseConn(ctx)
		return nil
	}

	if msg.Command == grpc.CmdHeartbeat {
		switch msg.PkgType {
		case grpc.PkgTypeRequest:
			//响应ping请求
			resp := &grpc.Message{
				Command: grpc.CmdHeartbeat,
				PkgType: grpc.PkgTypeReply,
				ReqId:   msg.ReqId,
				Seq:     msg.Seq,
				Result:  0,
			}
			return s.send(ctx, conn, resp)
		case grpc.PkgTypeReply:
			// TODO 记录上次ping响应时间
			return nil
		default:
			return nil
		}
	} else {
		reqKey := s.msgKey(msg)
		s.connManager.NewRequest(ctx, reqKey)
		go s.handler.Handle(msg, s.ch)
	}
	return nil
}

func (s *Server) prepare(conn netpoll.Connection) context.Context {
	ctx := context.Background()
	return ctx
}

func (s *Server) connect(ctx context.Context, conn netpoll.Connection) context.Context {
	connId := s.serialId.StringId()
	ctx = s.connManager.NewConn(ctx, connId, conn)
	_ = conn.AddCloseCallback(func(connection netpoll.Connection) error {
		s.connManager.CloseConn(ctx)
		return nil
	})

	return ctx
}

func (s *Server) send(ctx context.Context, conn netpoll.Connection, resp *grpc.Message) error {
	encode, err := s.protocol.Encode(ctx, resp)
	if err != nil {
		s.logger.ErrorF(nil, "Send: encode fail, err=%+v", err)
		return nil
	}

	_, e := conn.Write(encode)
	if e != nil {
		s.connManager.CloseConn(ctx)
		return e
	}
	return nil
}

func (s *Server) msgKey(msg *grpc.Message) string {
	return fmt.Sprintf("%d_%d_%d", msg.Command, msg.ReqId, msg.Seq)
}

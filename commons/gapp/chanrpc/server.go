package chanrpc

import (
	"fmt"
	"github.com/qiafan666/gotato/commons/gapp/logger"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/gid"
	"reflect"
	"runtime"
	"time"
)

// Server 代理服务器
type Server struct {
	handlers map[uint32]Handler
	chanReq  chan *ReqCtx
}

// NewServer 新建服务器
func NewServer(l int) *Server {
	s := new(Server)
	s.handlers = make(map[uint32]Handler)
	s.chanReq = make(chan *ReqCtx, l)
	return s
}

// Register 注册处理函数
func (s *Server) Register(msg any, f Handler) {
	msgID := MsgID(msg)
	if _, ok := s.handlers[msgID]; ok {
		panic(fmt.Sprintf("chanrpc Server Register msg ID %v: already registered", reflect.TypeOf(msg)))
	}
	s.handlers[msgID] = f
}

// RegisterByName 通过消息名注册消息处理函数
func (s *Server) RegisterByName(msgName string, f Handler) {
	msgID := gcommon.Str2Uint32(msgName)
	if _, ok := s.handlers[msgID]; ok {
		panic(fmt.Sprintf("chanrpc Server RegisterByName msg ID %v: already registered", msgID))
	}
	s.handlers[msgID] = f
}

// Exist 是否存在消息处理函数
func (s *Server) Exist(msg any) bool { return s.handlers[MsgID(msg)] != nil }

// exec 实际执行
func (s *Server) exec(reqCtx *ReqCtx) (err error) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 2048)
			l := runtime.Stack(buf, false)
			str := string(buf[:l])
			err = fmt.Errorf("%v: %s", r, str)
			// 如果是Cast，那么消息不会被返回
			reqCtx.ReplyErr(fmt.Errorf("%w: %v", ErrPeerPanic, r))
		}
	}()
	// 根据id取handler
	handler, ok := s.handlers[reqCtx.id]
	if !ok {
		err = fmt.Errorf("%w: msgType: %v, msgID: %v", ErrPeerMsgNotRegsitered, reflect.TypeOf(reqCtx.Req), reqCtx.id)
		reqCtx.ReplyErr(err)
		return err
	}
	handler(reqCtx)
	return err
}

// Len 当前任务队列长度
func (s *Server) Len() int {
	return len(s.chanReq)
}

// Exec 执行
func (s *Server) Exec(reqCtx *ReqCtx) {
	reqCtx.replied = false
	err := s.exec(reqCtx)
	if err != nil {
		logger.DefaultLogger.ErrorF("chanrpc Server Exec error: %v", err)
	}
}

// Cast 异步投递消息
func (s *Server) Cast(req any) {
	id := MsgID(req)
	reqCtx := &ReqCtx{
		reqID: gid.ID(),
		id:    id,
		Req:   req,
	}
	s.PendReq(reqCtx, false)
}

// Call 发起同步调用，不带超时机制
func (s *Server) Call(req any) *AckCtx {
	return NewClient(0).Call(s, req)
}

// CallT 发起同步调用，带超时机制
func (s *Server) CallT(req any, timeout time.Duration) *AckCtx {
	return NewClient(0).CallT(s, req, timeout)
}

// ChanReq 返回Call Channel
func (s *Server) ChanReq() chan *ReqCtx {
	return s.chanReq
}

// PendReq 将请求放入请求通道，当channel full时，返回ErrPeerChanRPCFull
func (s *Server) PendReq(reqCtx *ReqCtx, block bool) {
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			reqCtx.ReplyErr(err)
		}
	}()

	if block { // 阻塞模式
		s.chanReq <- reqCtx
		return
	}

	// 非阻塞模式
	select {
	case s.chanReq <- reqCtx:
	default:
		reqCtx.ReplyErr(ErrPeerChanRPCFull)
	}
}

// Close 关闭服务器
func (s *Server) Close() {
	close(s.chanReq)
	for reqCtx := range s.chanReq {
		reqCtx.ReplyErr(ErrPeerChanRPCClosed)
	}
}

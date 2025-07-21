package grpc

import "github.com/qiafan666/gotato/commons/gface"

type Handler interface {
	Handle(*Message) *Message
}

type Router struct {
	routes map[Command]Handler
	logger gface.ILogger
}

func NewRouter(logger gface.ILogger) *Router {
	r := &Router{routes: make(map[Command]Handler), logger: logger}
	return r
}

func (r *Router) Register(cmd Command, handler Handler) {
	r.routes[cmd] = handler
}

func (r *Router) Handle(msg *Message, out chan<- *Message) {
	r.logger.DebugF(nil, "grpc handle request msg, command:%d,reqId:%d,data:%s", msg.Command, msg.ReqId, msg.Body)
	if handler, ok := r.routes[msg.Command]; ok {
		resp := handler.Handle(msg)
		r.logger.DebugF(nil, "grpc handle response msg, command:%d,reqId:%d,data:%s", resp.Command, resp.ReqId, resp.Body)
		out <- resp
	} else {
		r.logger.ErrorF(nil, "grpc handle request msg:%v, command not found", msg)
	}
}

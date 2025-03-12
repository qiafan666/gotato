package grpc

import (
	"context"
)

type Server interface {
	Run(context.Context)
}

type Handler interface {
	Handle(request *Message, ch chan<- *Message)
}

// SyncHandler 同步处理
type SyncHandler interface {
	Handler
	Run(context.Context)
}

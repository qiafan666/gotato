package grpc

import (
	"context"
)

type IServer interface {
	Run(context.Context)
}

type IHandler interface {
	Handle(request *Message, ch chan<- *Message)
}

// SyncHandler 同步处理
type SyncHandler interface {
	IHandler
	Run(context.Context)
}

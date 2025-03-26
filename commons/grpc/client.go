package grpc

import (
	"context"
)

// IClient grpc 客户端调用接口
type IClient interface {
	Do(ctx context.Context, request *Message) (*Message, error)
}

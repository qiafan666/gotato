package grpc

import (
	"context"
)

type Client interface {
	Do(ctx context.Context, request *Message) (*Message, error)
}

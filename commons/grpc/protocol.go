package grpc

import (
	"context"
	"io"
)

type IProtocol interface {
	Encode(ctx context.Context, v *Message) ([]byte, error)
	Decode(ctx context.Context, reader io.Reader) (*Message, error)
}

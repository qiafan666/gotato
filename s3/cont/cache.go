package cont

import (
	"context"
	"github.com/qiafan666/gotato/s3"
)

type ICache interface {
	GetKey(ctx context.Context, engine string, key string) (*s3.ObjectInfo, error)
	DelS3Key(ctx context.Context, engine string, keys ...string) error
}

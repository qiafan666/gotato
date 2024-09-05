package cont

import (
	"context"
	"github.com/qiafan666/gotato/commons/gs3"
)

type S3Cache interface {
	GetKey(ctx context.Context, engine string, key string) (*gs3.ObjectInfo, error)
	DelS3Key(ctx context.Context, engine string, keys ...string) error
}

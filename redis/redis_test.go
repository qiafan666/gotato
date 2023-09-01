package redis

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/qiafan666/gotato"
	"time"
)

var Nil = errors.New("redis: nil")

type Dao interface {
	Client() *redis.Client
	Get(ctx context.Context, k string) (out string, err error)
	Set(ctx context.Context, k string, x interface{}, d time.Duration) (err error)
	Delete(ctx context.Context, k string) (err error)
}

type Imp struct {
	redis *redis.Client
}

func Instance() Dao {
	return &Imp{redis: gotato.GetGotatoInstance().Redis("test")}
}

func (i Imp) Client() *redis.Client {
	return i.redis
}

func (i Imp) Get(ctx context.Context, k string) (out string, err error) {
	result := i.redis.Get(ctx, k)
	if result.Err() != nil {
		return out, result.Err()
	}
	return result.Val(), nil
}

func (i Imp) Set(ctx context.Context, k string, x interface{}, d time.Duration) (err error) {
	marshal, err := json.Marshal(x)
	if err != nil {
		return err
	}
	return i.redis.Set(ctx, k, marshal, d).Err()
}

func (i Imp) Delete(ctx context.Context, k string) (err error) {
	return i.redis.Del(ctx, k).Err()
}

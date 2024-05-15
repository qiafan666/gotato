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

const (
	//没有过期时间
	NoExpiration = time.Duration(0)
	//ttl 保持现有过期时间
	KeepTTL = time.Duration(-1)
)

type Dao interface {
	Client() *redis.Client
	Get(ctx context.Context, k string) (out string, err error)
	Set(ctx context.Context, k string, x interface{}, d time.Duration) (err error)
	Delete(ctx context.Context, k string) (err error)
	Publish(ctx context.Context, channel string, message interface{}) (err error)
	Subscribe(ctx context.Context, channel string) (out <-chan *redis.Message, err error)
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

func (i Imp) Publish(ctx context.Context, channel string, message interface{}) (err error) {
	err = i.redis.Publish(ctx, channel, message).Err()
	if err != nil {
		return err
	}
	return nil
}

func (i Imp) Subscribe(ctx context.Context, channel string) (<-chan *redis.Message, error) {
	// 创建一个通道，用于发送接收到的消息
	out := make(chan *redis.Message)

	// 从 Redis 订阅给定的频道
	sub := i.redis.Subscribe(ctx, channel)
	// 使用 Go 协程来处理接收到的消息
	go func() {
		defer close(out) // 确保在协程结束时关闭通道
		for message := range sub.Channel() {
			// 将接收到的消息发送到通道
			out <- message
		}
	}()

	return out, nil
}

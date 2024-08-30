package redis

import (
	"context"
	"errors"
	"github.com/qiafan666/gotato/commons/gredis"
	"github.com/qiafan666/gotato/gconfig"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	redisSource redis.UniversalClient
	name        string //redis  name
}

func (slf *Redis) Redis() redis.UniversalClient {
	return slf.redisSource
}

func (slf *Redis) Name() string {
	return slf.name
}

func (slf *Redis) StartRedis(config gconfig.RedisConfig) error {

	err := gredis.Check(context.Background(), &gredis.Config{
		ClusterMode: config.ClusterMode,
		Address:     config.Address,
		Username:    config.Username,
		Password:    config.Password,
		MaxRetry:    config.MaxRetry,
		DB:          config.DB,
		PoolSize:    config.PoolSize,
	})
	if err != nil {
		return err
	}

	slf.name = config.Name
	client, err := gredis.NewRedisClient(context.Background(), &gredis.Config{
		ClusterMode: config.ClusterMode,
		Address:     config.Address,
		Username:    config.Username,
		Password:    config.Password,
		MaxRetry:    config.MaxRetry,
		DB:          config.DB,
		PoolSize:    config.PoolSize,
	})
	if err != nil {
		return err
	}
	slf.redisSource = client
	return nil
}

func (slf *Redis) StopRedis() error {
	if slf.redisSource == nil {
		return errors.New("redis not opened")
	}
	err := slf.redisSource.Close()
	if err != nil {
		slf.redisSource = nil
	}
	return err
}

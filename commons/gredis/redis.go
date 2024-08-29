package gredis

import (
	"context"
	"fmt"
	"github.com/qiafan666/gotato/commons/gerrs"
	"github.com/redis/go-redis/v9"
)

const defaultMaxRetry = 3
const defaultPoolSize = 100

// Config 连接配置参数，包括单机模式和集群模式的连接配置参数。
type Config struct {
	ClusterMode bool     // 是否启用集群模式。
	Address     []string // 地址
	Username    string   // 用户名
	Password    string   // 密码
	MaxRetry    int      // 最大重试次数。
	DB          int      // 数据库编号。
	PoolSize    int      // 连接池大小。
}

func NewRedisClient(ctx context.Context, config *Config) (redis.UniversalClient, error) {
	if len(config.Address) == 0 {
		return nil, gerrs.New("redis address is empty")
	}

	if config.MaxRetry == 0 {
		config.MaxRetry = defaultMaxRetry
	}
	if config.PoolSize == 0 {
		config.PoolSize = defaultPoolSize
	}

	var cli redis.UniversalClient
	if config.ClusterMode || len(config.Address) > 1 {
		opt := &redis.ClusterOptions{
			Addrs:      config.Address,
			Username:   config.Username,
			Password:   config.Password,
			PoolSize:   config.PoolSize,
			MaxRetries: config.MaxRetry,
		}
		cli = redis.NewClusterClient(opt)
	} else {
		opt := &redis.Options{
			Addr:       config.Address[0],
			Username:   config.Username,
			Password:   config.Password,
			DB:         config.DB,
			PoolSize:   config.PoolSize,
			MaxRetries: config.MaxRetry,
		}
		cli = redis.NewClient(opt)
	}
	if err := cli.Ping(ctx).Err(); err != nil {
		return nil, gerrs.New(fmt.Sprintf("Redis Ping failed: %v, Address: %v, Username: %v, ClusterMode: %v", err, config.Address, config.Username, config.ClusterMode))
	}
	return cli, nil
}

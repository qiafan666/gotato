package gredis

import (
	"context"
	"errors"
	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"
	"time"
)

const defaultMaxRetry = 3
const defaultPoolSize = 100

var json = jsoniter.ConfigCompatibleWithStandardLibrary

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

func NewRedisClient(ctx context.Context, config *Config) (*Client, error) {
	if len(config.Address) == 0 {
		return nil, errors.New("redis address is empty")
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
		return nil, err
	}
	return &Client{redis: cli}, nil
}

type Client struct {
	redis redis.UniversalClient
}

// GetRedis 获取 Redis 客户端。
func (c *Client) GetRedis() redis.UniversalClient {
	return c.redis
}

// Set 将一个键值对存储到 Redis，值会被序列化为 JSON 格式。
// 返回一个布尔值表示是否成功，以及可能发生的错误。
func (c *Client) Set(key string, val interface{}) (bool, error) {
	marshal, err := json.Marshal(val)
	if err != nil {
		return false, err
	}
	status := c.redis.Set(context.Background(), key, marshal, 0)
	if status.Err() != nil {
		return false, status.Err()
	}
	return true, nil
}

// Del 从 Redis 中删除一个键。
// 返回一个布尔值表示是否成功，以及可能发生的错误。
func (c *Client) Del(key string) (bool, error) {
	status := c.redis.Del(context.Background(), key)
	if status.Err() != nil {
		return false, status.Err()
	}
	return true, nil
}

// SetEx 将一个键值对存储到 Redis，并设置过期时间，值会被序列化为 JSON 格式。
// 返回一个布尔值表示是否成功，以及可能发生的错误。
func (c *Client) SetEx(key string, val interface{}, expire time.Duration) (bool, error) {
	marshal, err := json.Marshal(val)
	if err != nil {
		return false, err
	}
	status := c.redis.SetEx(context.Background(), key, marshal, expire)
	if status.Err() != nil {
		return false, status.Err()
	}
	return true, nil
}

// Get 从 Redis 获取一个键的值，并将其反序列化为指定的类型。
// 返回可能发生的错误。
func (c *Client) Get(key string, val interface{}) error {
	cmd := c.redis.Get(context.Background(), key)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	bytes, _ := cmd.Bytes()
	err := json.Unmarshal(bytes, val)
	return err
}

// Keys 根据给定的模式获取所有匹配的键。
// 返回匹配的键列表，以及可能发生的错误。
func (c *Client) Keys(pattern string) ([]string, error) {
	cmd := c.redis.Keys(context.Background(), pattern)
	if cmd.Err() != nil {
		return []string{}, nil
	}
	return cmd.Result()
}

// HVals 获取哈希键所有字段的值。
// 返回哈希键的值列表，以及可能发生的错误。
func (c *Client) HVals(key string) ([]string, error) {
	vals := c.redis.HVals(context.Background(), key)
	return vals.Val(), vals.Err()
}

// HSet 向哈希表中设置字段值。支持以下几种传参方式：
//   - HSet("myhash", "key1", "value1", "key2", "value2")
//   - HSet("myhash", []string{"key1", "value1", "key2", "value2"})
//   - HSet("myhash", map[string]interface{}{"key1": "value1", "key2": "value2"})
//
// 返回一个布尔值表示是否成功，以及可能发生的错误。
func (c *Client) HSet(key string, values ...interface{}) (bool, error) {
	status := c.redis.HSet(context.Background(), key, values...)
	if status.Err() != nil {
		return false, status.Err()
	}
	return true, nil
}

// HGet 获取哈希表中指定字段的值。
// 返回字段的值和可能发生的错误。
func (c *Client) HGet(key, field string) (string, error) {
	cmd := c.redis.HGet(context.Background(), key, field)
	if cmd.Err() != nil {
		return "", cmd.Err()
	}
	return cmd.Result()
}

// HGetAll 获取哈希表中所有字段的值。
// 返回哈希表的所有字段及值，以及可能发生的错误。
func (c *Client) HGetAll(key string) (map[string]string, error) {
	cmd := c.redis.HGetAll(context.Background(), key)
	if cmd.Err() != nil {
		return nil, cmd.Err()
	}
	return cmd.Result()
}

// ZAdd 向有序集合中添加一个或多个成员。
// 返回一个布尔值表示是否成功，以及可能发生的错误。
func (c *Client) ZAdd(key string, slice []redis.Z) (bool, error) {
	status := c.redis.ZAdd(context.Background(), key, slice...)
	if status.Err() != nil {
		return false, status.Err()
	}
	return true, nil
}

// ZRevRangeByScoreWithScores 根据分数范围逆序获取有序集合中的成员及分数。
// 返回成员及分数的列表，以及可能发生的错误。
func (c *Client) ZRevRangeByScoreWithScores(key string, opt *redis.ZRangeBy) ([]redis.Z, error) {
	res, err := c.redis.ZRevRangeByScoreWithScores(context.Background(), key, opt).Result()
	return res, err
}

// ZRemRangeByScore 删除有序集合中指定分数范围的成员。
// 返回布尔值表示是否成功，以及可能发生的错误。
func (c *Client) ZRemRangeByScore(key, min, max string) (bool, error) {
	status := c.redis.ZRemRangeByScore(context.Background(), key, min, max)
	if err := status.Err(); err != nil {
		return false, err
	}
	return true, nil
}

// ZRevRangeWithScores 根据排名逆序获取有序集合中的成员及分数。
// 返回成员及分数的列表，以及可能发生的错误。
func (c *Client) ZRevRangeWithScores(key string, offset, limit int64) ([]redis.Z, error) {
	res, err := c.redis.ZRevRangeWithScores(context.Background(), key, offset, offset+limit-1).Result()
	return res, err
}

// ZRemRange 根据排名范围删除有序集合中的成员。
// 返回布尔值表示是否成功，以及可能发生的错误。
func (c *Client) ZRemRange(key string, start, stop int64) (bool, error) {
	status := c.redis.ZRemRangeByRank(context.Background(), key, start, stop)
	if err := status.Err(); err != nil {
		return false, err
	}
	return true, nil
}

// ZRem 删除有序集合中的指定成员。
// 返回布尔值表示是否成功，以及可能发生的错误。
func (c *Client) ZRem(key, member string) (bool, error) {
	status := c.redis.ZRem(context.Background(), key, member)
	if err := status.Err(); err != nil {
		return false, err
	}
	return true, nil
}

// Expire 设置指定键的过期时间。
// 返回可能发生的错误。
func (c *Client) Expire(key string, expiration time.Duration) error {
	status := c.redis.Expire(context.Background(), key, expiration)
	return status.Err()
}

// ZRangeByScoreWithScores 根据分数范围获取有序集合中的成员及分数。
// 返回成员及分数的列表，以及可能发生的错误。
func (c *Client) ZRangeByScoreWithScores(key string, min, max string) ([]redis.Z, error) {
	res, err := c.redis.ZRangeByScoreWithScores(context.Background(), key, &redis.ZRangeBy{
		Min: min,
		Max: max,
	}).Result()
	return res, err
}

// ZRangeByScoreAndLimitWithScores 根据分数范围和限制条件获取有序集合中的成员及分数。
// 返回成员及分数的列表，以及可能发生的错误。
func (c *Client) ZRangeByScoreAndLimitWithScores(key string, min, max string, limit int64) ([]redis.Z, error) {
	res, err := c.redis.ZRangeByScoreWithScores(context.Background(), key, &redis.ZRangeBy{
		Min:   min,
		Max:   max,
		Count: limit,
	}).Result()
	return res, err
}

// RPush 将一个值推送到 Redis 列表的右端，值会被序列化为 JSON 格式。
// 返回布尔值表示是否成功，以及可能发生的错误。
func (c *Client) RPush(key string, val interface{}) (bool, error) {
	marshal, err := json.Marshal(val)
	if err != nil {
		return false, err
	}
	status := c.redis.RPush(context.Background(), key, marshal)
	if status.Err() != nil {
		return false, status.Err()
	}
	return true, nil
}

// RPushStr 将一个或多个字符串值推送到 Redis 列表的右端。
// 返回布尔值表示是否成功，以及可能发生的错误。
func (c *Client) RPushStr(key string, val ...string) (bool, error) {
	status := c.redis.RPush(context.Background(), key, val)
	if status.Err() != nil {
		return false, status.Err()
	}
	return true, nil
}

// ZRevRank 返回成员在有序集合中的排名，从 0 开始，若成员不存在，则返回 -1。
// 返回成员的排名，以及可能发生的错误。
func (c *Client) ZRevRank(key string, member string) (int64, error) {
	rank, err := c.redis.ZRevRank(context.Background(), key, member).Result()
	if err != nil {
		return -1, nil
	}
	return rank, nil
}

// IncrBy 自增指定键的值，返回自增后的值。
// 返回自增后的结果，以及可能发生的错误。
func (c *Client) IncrBy(key string, incr int64) (int64, error) {
	result, err := c.redis.IncrBy(context.Background(), key, incr).Result()
	if err != nil && !errors.Is(redis.Nil, err) {
		return 0, err
	}
	return result, nil
}

// SetBit 设置指定偏移量的位值。
func (c *Client) SetBit(key string, offset int64, value int) (bool, error) {
	status := c.redis.SetBit(context.Background(), key, offset, value)
	if status.Err() != nil {
		// 如果有错误，输出日志
		return false, status.Err()
	}
	return true, nil
}

// GetBit 获取指定偏移量的位值。
func (c *Client) GetBit(key string, offset int64) (int, error) {
	result, err := c.redis.GetBit(context.Background(), key, offset).Result()
	if err != nil {
		// 如果有错误，输出日志
		return 0, err
	}
	return int(result), nil
}

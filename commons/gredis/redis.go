package gredis

import (
	"context"
	"errors"
	"github.com/qiafan666/gotato/commons/gson"
	"github.com/redis/go-redis/v9"
	"time"
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

func NewClient(ctx context.Context, config *Config) (*Client, error) {
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

func SetRedis(redis redis.UniversalClient) *Client {
	return &Client{redis: redis}
}

type Client struct {
	redis redis.UniversalClient
}

// GetRedis 获取 Redis 客户端。
func (c *Client) GetRedis() redis.UniversalClient {
	return c.redis
}

// Close 关闭 Redis 连接。
func (c *Client) Close() error {
	return c.redis.Close()
}

// Set 将一个键值对存储到 Redis，值会被序列化为 JSON 格式。
// 返回一个布尔值表示是否成功，以及可能发生的错误。
func (c *Client) Set(key string, val interface{}) (bool, error) {
	marshal, err := gson.Marshal(val)
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
	marshal, err := gson.Marshal(val)
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
	err := gson.Unmarshal(bytes, val)
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
	marshal, err := gson.Marshal(val)
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

// RListRange 返回 Redis 列表中指定范围内的元素
func (c *Client) RListRange(key string, start, stop int64) ([]string, error) {
	result, err := c.redis.LRange(context.Background(), key, start, stop).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// RListRemove 从 Redis 列表中删除指定的多个元素（事务方式）
func (c *Client) RListRemove(key string, values ...string) (int64, error) {
	// Lua 脚本
	script := `
        local key = KEYS[1]
        local count = 0
        for i = 1, #ARGV do
            count = count + redis.call('LREM', key, 0, ARGV[i])
        end
        return count
    `

	// 将 values 转换为 []interface{}
	args := make([]interface{}, len(values))
	for i, v := range values {
		args[i] = v
	}

	// 执行 Lua 脚本
	result, err := c.redis.Eval(context.Background(), script, []string{key}, args...).Result()
	if err != nil {
		return 0, err
	}

	// 返回删除的元素数量
	return result.(int64), nil
}

// RListIsContain 判断元素是否在 Redis 列表中
func (c *Client) RListIsContain(key string, target string) (bool, error) {
	// 获取列表中的所有元素
	elements, err := c.redis.LRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		return false, err
	}

	// 遍历列表，检查是否存在目标元素
	for _, element := range elements {
		if element == target {
			return true, nil
		}
	}

	return false, nil
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

// DecrBy 自减指定键的值，返回自减后的值。
// 返回自减后的结果，以及可能发生的错误。
func (c *Client) DecrBy(key string, decr int64) (int64, error) {
	result, err := c.redis.DecrBy(context.Background(), key, decr).Result()
	if err != nil && !errors.Is(redis.Nil, err) {
		return 0, err
	}
	return result, nil
}

// SAdd 将一个或多个元素添加到集合中。
// 返回布尔值表示是否成功，以及可能发生的错误。
func (c *Client) SAdd(key string, members ...interface{}) (bool, error) {
	status := c.redis.SAdd(context.Background(), key, members...)
	if status.Err() != nil {
		return false, status.Err()
	}
	return true, nil
}

// SRem 从集合中删除一个或多个元素。
// 返回布尔值表示是否成功，以及可能发生的错误。
func (c *Client) SRem(key string, members ...interface{}) (bool, error) {
	status := c.redis.SRem(context.Background(), key, members...)
	if status.Err() != nil {
		return false, status.Err()
	}
	return true, nil
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

// BitCount 统计位图中值为 1 的位的数量。
func (c *Client) BitCount(key string, bitCount *redis.BitCount) (int64, error) {
	result, err := c.redis.BitCount(context.Background(), key, bitCount).Result()
	if err != nil {
		// 如果有错误，输出日志
		return 0, err
	}
	return result, nil
}

// GetBits 批量获取位图的值
func (c *Client) GetBits(key string, start, size int64) ([]int64, error) {
	// 使用 BITFIELD 命令批量获取位的值
	result, err := c.redis.BitField(context.Background(), key, "GET", "u1", start, "GET", "u1", start+1, "GET", "u1", start+size-1).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// SetNextBit 根据规则设置位图的值,返回设置的偏移量。
func (c *Client) SetNextBit(key string, batchSize int64) (int64, error) {
	var offset int64 = 0
	for {
		// 批量获取位的值
		bits, err := c.GetBits(key, offset, batchSize)
		if err != nil {
			return 0, err
		}

		// 查找第一个值为 0 的位
		for i, bit := range bits {
			if bit == 0 {
				targetOffset := offset + int64(i)
				// 使用 Lua 脚本确保原子性
				script := `
                    local key = KEYS[1]
                    local offset = tonumber(ARGV[1])
                    local bit = redis.call('GETBIT', key, offset)
                    if bit == 0 then
                        redis.call('SETBIT', key, offset, 1)
                        return offset
                    else
                        return -1
                    end
                `
				result, err := c.redis.Eval(context.Background(), script, []string{key}, targetOffset).Result()
				if err != nil {
					return 0, err
				}
				if result.(int64) != -1 {
					return result.(int64), nil
				}
			}
		}

		// 如果没有找到，继续查询下一个范围
		offset += batchSize
	}
}

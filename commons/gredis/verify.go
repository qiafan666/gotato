package gredis

import (
	"context"
)

// Check 检查Redis连接是否正常
func Check(ctx context.Context, config *Config) error {
	client, err := NewClient(ctx, config)
	if err != nil {
		return err
	}
	defer client.Close()

	// Ping the Redis server to check connectivity.
	if err = client.GetRedis().Ping(ctx).Err(); err != nil {
		return err
	}

	return nil
}

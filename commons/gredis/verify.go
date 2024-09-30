package gredis

import (
	"context"
)

// Check 检查Redis连接是否正常
func Check(ctx context.Context, config *Config) error {
	client, err := NewRedisClient(ctx, config)
	if err != nil {
		return err
	}
	defer client.Close()

	// Ping the Redis server to check connectivity.
	if err = client.Ping(ctx).Err(); err != nil {
		return err
	}

	return nil
}

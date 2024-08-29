package gredis

import (
	"context"
	"errors"
	"fmt"
	"github.com/qiafan666/gotato/commons/gcast"
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
		return errors.New(fmt.Sprintf("Failed to ping Redis server: %s,config: %s", err, gcast.ToString(config)))
	}

	return nil
}

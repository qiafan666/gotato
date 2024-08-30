package gmongo

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"strings"
)

const (
	defaultMaxPoolSize = 100
	defaultMaxRetry    = 3
)

// buildMongoURI constructs the MongoDB URI from the provided configuration.
func buildMongoURI(config *Config) string {
	credentials := ""
	if config.Username != "" && config.Password != "" {
		credentials = fmt.Sprintf("%s:%s@", config.Username, config.Password)
	}
	return fmt.Sprintf("mongodb://%s%s/%s?maxPoolSize=%d", credentials, strings.Join(config.Address, ","), config.Database, config.MaxPoolSize)
}

// shouldRetry determines whether an error should trigger a retry.
func shouldRetry(ctx context.Context, err error) bool {
	select {
	case <-ctx.Done():
		return false
	default:
		var cmdErr mongo.CommandError
		if errors.As(err, &cmdErr) {
			return cmdErr.Code != 13 && cmdErr.Code != 18
		}
		return true
	}
}

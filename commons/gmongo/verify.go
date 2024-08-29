package gmongo

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Check 检查配置是否正确
func Check(ctx context.Context, config *Config) error {
	if err := config.ValidateAndSetDefaults(); err != nil {
		return err
	}

	clientOpts := options.Client().ApplyURI(config.Uri)
	mongoClient, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return errors.Wrapf(err, "MongoDB connect failed, URI: %s, Database: %s, MaxPoolSize: %d", config.Uri, config.Database, config.MaxPoolSize)
	}

	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			_ = mongoClient.Disconnect(ctx)
		}
	}()

	if err = mongoClient.Ping(ctx, nil); err != nil {
		return errors.Wrapf(err, "MongoDB ping failed, URI: %s, Database: %s, MaxPoolSize: %d", config.Uri, config.Database, config.MaxPoolSize)
	}

	return nil
}

// ValidateAndSetDefaults 验证配置并设置默认值
func (c *Config) ValidateAndSetDefaults() error {
	if c.Uri == "" && len(c.Address) == 0 {
		return errors.New("either Uri or Address must be provided")
	}
	if c.Database == "" {
		return errors.New("database is required")
	}
	if c.MaxPoolSize <= 0 {
		c.MaxPoolSize = defaultMaxPoolSize
	}
	if c.MaxRetry <= 0 {
		c.MaxRetry = defaultMaxRetry
	}
	if c.Uri == "" {
		c.Uri = buildMongoURI(c)
	}
	return nil
}

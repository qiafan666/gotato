package gmongo

import (
	"context"
	"github.com/qiafan666/gotato/commons/gmongo/tx"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// Config represents the MongoDB configuration.
type Config struct {
	Uri         string //填了uri可以不用填其他参数
	Address     []string
	Database    string
	Username    string
	Password    string
	MaxPoolSize int
	MaxRetry    int
}

type Client struct {
	tx tx.Tx
	db *mongo.Database
}

func (c *Client) GetDB() *mongo.Database {
	return c.db
}

func (c *Client) GetTx() tx.Tx {
	return c.tx
}

// NewMongoDB initializes a new MongoDB connection.
func NewMongoDB(ctx context.Context, config *Config) (*Client, error) {
	if err := config.ValidateAndSetDefaults(); err != nil {
		return nil, err
	}
	opts := options.Client().ApplyURI(config.Uri).SetMaxPoolSize(uint64(config.MaxPoolSize))
	var (
		cli *mongo.Client
		err error
	)
	for i := 0; i < config.MaxRetry; i++ {
		cli, err = connectMongo(ctx, opts)
		if err != nil && shouldRetry(ctx, err) {
			time.Sleep(time.Second / 2)
			continue
		}
		break
	}
	if err != nil {
		return nil, err
	}
	mtx, err := NewMongoTx(ctx, cli)
	if err != nil {
		return nil, err
	}
	return &Client{
		tx: mtx,
		db: cli.Database(config.Database),
	}, nil
}

func connectMongo(ctx context.Context, opts *options.ClientOptions) (*mongo.Client, error) {
	cli, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}
	if err = cli.Ping(ctx, nil); err != nil {
		return nil, err
	}
	return cli, nil
}

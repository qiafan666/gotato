package nats

import (
	"context"
	"log"

	"github.com/nats-io/nats.go"
)

type clientImpl struct {
	conn *nats.Conn
}

// INat 定义 NATS 客户端接口
type INat interface {
	Publish(subject string, data []byte) error
	Subscribe(subject string, callback func(msg *nats.Msg)) (*nats.Subscription, error)
	Unsubscribe(sub *nats.Subscription) error
	Request(ctx context.Context, subject string, data []byte) (*nats.Msg, error)
	Reply(subject string, reply string, data []byte) error
	RequestMsg(ctx context.Context, msg *nats.Msg) (*nats.Msg, error)
	Flush() error
	Connection() *nats.Conn
	Close()
}

// NewNATSClient 创建新的 NATS 客户端实例
func NewNATSClient(url string, opts ...nats.Option) (INat, error) {
	conn, err := nats.Connect(url, opts...)
	if err != nil {
		log.Printf("Error connecting to NATS: %v", err)
		return nil, err
	}
	return &clientImpl{conn: conn}, nil
}

// Publish 发送消息到指定的主题
func (c *clientImpl) Publish(subject string, data []byte) error {
	return c.conn.Publish(subject, data)
}

// Subscribe 订阅指定主题
func (c *clientImpl) Subscribe(subject string, callback func(msg *nats.Msg)) (*nats.Subscription, error) {
	return c.conn.Subscribe(subject, callback)
}

// Unsubscribe 取消订阅
func (c *clientImpl) Unsubscribe(sub *nats.Subscription) error {
	return sub.Unsubscribe()
}

// Request 发送请求并等待响应
func (c *clientImpl) Request(ctx context.Context, subject string, data []byte) (*nats.Msg, error) {
	return c.conn.RequestWithContext(ctx, subject, data)
}

// RequestMsg 发送请求消息并等待响应
func (c *clientImpl) RequestMsg(ctx context.Context, msg *nats.Msg) (*nats.Msg, error) {
	return c.conn.RequestMsgWithContext(ctx, msg)
}

// Reply 发送响应消息到指定的主题和回复地址
func (c *clientImpl) Reply(subject string, reply string, data []byte) error {
	return c.conn.PublishRequest(subject, reply, data)
}

// Flush 刷新连接，确保所有数据都已发送
func (c *clientImpl) Flush() error {
	return c.conn.Flush()
}

// Connection 返回当前 NATS 连接
func (c *clientImpl) Connection() *nats.Conn {
	return c.conn
}

// Close 关闭 NATS 连接并确保在关闭前刷新连接
func (c *clientImpl) Close() {
	err := c.conn.Flush()
	if err != nil {
		log.Printf("Error flushing connection: %v", err)
	}
	c.conn.Close()
}

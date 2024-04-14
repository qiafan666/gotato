package nats

import (
	"context"
	"log"

	"github.com/nats-io/nats.go"
)

type ClientImpl struct {
	conn *nats.Conn
}

// Nats 定义 NATS 客户端接口
type Nats interface {
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
func NewNATSClient(url string, opts ...nats.Option) (*ClientImpl, error) {
	conn, err := nats.Connect(url, opts...)
	if err != nil {
		log.Printf("Error connecting to NATS: %v", err)
		return nil, err
	}
	return &ClientImpl{conn: conn}, nil
}

// Publish 发送消息到指定的主题
func (n *ClientImpl) Publish(subject string, data []byte) error {
	return n.conn.Publish(subject, data)
}

// Subscribe 订阅指定主题
func (n *ClientImpl) Subscribe(subject string, callback func(msg *nats.Msg)) (*nats.Subscription, error) {
	return n.conn.Subscribe(subject, callback)
}

// Unsubscribe 取消订阅
func (n *ClientImpl) Unsubscribe(sub *nats.Subscription) error {
	return sub.Unsubscribe()
}

// Request 发送请求并等待响应
func (n *ClientImpl) Request(ctx context.Context, subject string, data []byte) (*nats.Msg, error) {
	return n.conn.RequestWithContext(ctx, subject, data)
}

// RequestMsg 发送请求消息并等待响应
func (n *ClientImpl) RequestMsg(ctx context.Context, msg *nats.Msg) (*nats.Msg, error) {
	return n.conn.RequestMsgWithContext(ctx, msg)
}

// Reply 发送响应消息到指定的主题和回复地址
func (n *ClientImpl) Reply(subject string, reply string, data []byte) error {
	return n.conn.PublishRequest(subject, reply, data)
}

// Flush 刷新连接，确保所有数据都已发送
func (n *ClientImpl) Flush() error {
	return n.conn.Flush()
}

// Connection 返回当前 NATS 连接
func (n *ClientImpl) Connection() *nats.Conn {
	return n.conn
}

// Close 关闭 NATS 连接并确保在关闭前刷新连接
func (n *ClientImpl) Close() {
	err := n.conn.Flush()
	if err != nil {
		log.Printf("Error flushing connection: %v", err)
	}
	n.conn.Close()
}

package fanout

import (
	"context"
	"time"
)

type ConsumerConfig struct {
	Ctx context.Context
	// Addr  地址
	Addr string
	// QueueName  队列名称
	QueueName string
	// Durable  是否持久化
	Durable bool
	// ChanNumber  消费者数量
	ChanNumber int
	// ReconnectInterval  重连间隔
	ReconnectInterval time.Duration
	// RetryTimes  重试次数
	RetryTimes int
	// ExchangeType  exchange类型
	ExchangeType string
	// ExchangeName  exchange名称
	ExchangeName string
	// DelayExchangeName  延迟exchange名称
	DelayExchangeName string
}

type ProducerConfig struct {
	Ctx context.Context
	// Addr  地址
	Addr string
	// QueueName  队列名称
	QueueName string
	// Durable  是否持久化
	Durable bool
	// ExchangeType  exchange类型
	ExchangeType string
	// ExchangeName  exchange名称
	ExchangeName string
}

package work_queue

import (
	"context"
	"time"
)

type ConsumerConfig struct {
	Ctx context.Context
	// RabbitMqAddr rabbitmq 地址
	Addr string
	// RabbitMqQueueName rabbitmq 队列名称
	QueueName string
	// RabbitMqDurable rabbitmq 是否持久化
	Durable bool
	// RabbitMqChanNumber rabbitmq 通道数量
	ChanNumber int
	// RabbitMqReconnectInterval rabbitmq 重连间隔
	ReconnectInterval time.Duration
	// RabbitMqRetryTimes rabbitmq 重试次数
	RetryTimes int
}

type ProducerConfig struct {
	Ctx context.Context
	// RabbitMqAddr rabbitmq 地址
	Addr string
	// RabbitMqQueueName rabbitmq 队列名称
	QueueName string
	// RabbitMqDurable rabbitmq 是否持久化
	Durable bool
}

package hello_world

import "time"

type ConsumerConfig struct {
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
	// RabbitMqAddr rabbitmq 地址
	Addr string
	// RabbitMqQueueName rabbitmq 队列名称
	QueueName string
	// RabbitMqDurable rabbitmq 是否持久化
	Durable bool
}

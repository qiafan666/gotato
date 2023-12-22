package direct

import "time"

type ConsumerConfig struct {
	// RabbitMqAddr rabbitmq 地址
	Addr string
	// RabbitMqQueueName rabbitmq 队列名称
	QueueName string
	// RabbitMqDurable rabbitmq 是否持久化
	Durable bool
	// RabbitMqReconnectInterval rabbitmq 重连间隔
	ReconnectInterval time.Duration
	// RabbitMqRetryTimes rabbitmq 重试次数
	RetryTimes int
	// RabbitMqExchangeType rabbitmq exchange类型
	ExchangeType string
	// RabbitMqExchangeName rabbitmq exchange名称
	ExchangeName string
	// RabbitMqDelayExchangeName rabbitmq 延迟exchange名称
	DelayExchangeName string
}

type ProducerConfig struct {
	// RabbitMqAddr rabbitmq 地址
	Addr string
	// RabbitMqQueueName rabbitmq 队列名称
	QueueName string
	// RabbitMqDurable rabbitmq 是否持久化
	Durable bool
	// RabbitMqExchangeType rabbitmq exchange类型
	ExchangeType string
	// RabbitMqExchangeName rabbitmq exchange名称
	ExchangeName string
}

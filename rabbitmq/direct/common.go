package direct

import "time"

var (
	// RabbitMqAddr rabbitmq 地址
	RabbitMqAddr string
	// RabbitMqQueueName rabbitmq 队列名称
	RabbitMqQueueName string
	// RabbitMqDurable rabbitmq 是否持久化
	RabbitMqDurable bool
	// RabbitMqReconnectInterval rabbitmq 重连间隔
	RabbitMqReconnectInterval time.Duration
	// RabbitMqRetryTimes rabbitmq 重试次数
	RabbitMqRetryTimes int
	// RabbitMqExchangeType rabbitmq exchange类型
	RabbitMqExchangeType string
	// RabbitMqExchangeName rabbitmq exchange名称
	RabbitMqExchangeName string
	// RabbitMqDelayExchangeName rabbitmq 延迟exchange名称
	RabbitMqDelayExchangeName string
)

func StartDirect(addr string, queueName string, durable bool, reconnectInterval time.Duration, retryTimes int, exchangeType string, exchangeName string, delayExchangeName string) {

	RabbitMqAddr = addr
	RabbitMqQueueName = queueName
	RabbitMqDurable = durable
	RabbitMqReconnectInterval = reconnectInterval
	RabbitMqRetryTimes = retryTimes
	RabbitMqExchangeType = exchangeType
	RabbitMqExchangeName = exchangeName
	RabbitMqDelayExchangeName = delayExchangeName
}

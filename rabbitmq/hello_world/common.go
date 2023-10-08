package hello_world

import "time"

var (
	// RabbitMqAddr rabbitmq 地址
	RabbitMqAddr string
	// RabbitMqQueueName rabbitmq 队列名称
	RabbitMqQueueName string
	// RabbitMqDurable rabbitmq 是否持久化
	RabbitMqDurable bool
	// RabbitMqChanNumber rabbitmq 通道数量
	RabbitMqChanNumber int
	// RabbitMqReconnectInterval rabbitmq 重连间隔
	RabbitMqReconnectInterval time.Duration
	// RabbitMqRetryTimes rabbitmq 重试次数
	RabbitMqRetryTimes int
)

func StartHelloWorld(addr string, queueName string, durable bool, chanNumber int, reconnectInterval time.Duration, retryTimes int, exchangeType string, exchangeName string, delayExchangeName string) {

	RabbitMqAddr = addr
	RabbitMqQueueName = queueName
	RabbitMqDurable = durable
	RabbitMqChanNumber = chanNumber
	RabbitMqReconnectInterval = reconnectInterval
	RabbitMqRetryTimes = retryTimes

}

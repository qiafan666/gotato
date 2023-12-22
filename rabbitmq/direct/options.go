package direct

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// 目前 apply(*producer) 的参数只能固定为生产者或者消费者其中之一的具体类型
// 1.生产者初始化参数定义

// OptionsProd 定义动态设置参数接口
type OptionsProd interface {
	apply(*Producer)
}

// OptionFunc 以函数形式实现上面的接口
type OptionFunc func(*Producer)

func (f OptionFunc) apply(prod *Producer) {
	f(prod)
}

// SetProdMsgDelayParams 开发者设置生产者初始化时的参数
func SetProdMsgDelayParams(delayExchangeName string, enableMsgDelayPlugin bool) OptionsProd {
	return OptionFunc(func(p *Producer) {
		p.enableDelayMsgPlugin = enableMsgDelayPlugin
		p.config.ExchangeType = "x-delayed-message"
		p.args = amqp.Table{
			"x-delayed-type": "direct",
		}
		p.config.ExchangeName = delayExchangeName
		// 延迟消息队列，交换机、消息全部设置为持久
		p.config.Durable = true
	})
}

// 2.消费者端初始化参数定义

// OptionsConsumer 定义动态设置参数接口
type OptionsConsumer interface {
	apply(*Consumer)
}

// OptionsConsumerFunc 以函数形式实现上面的接口
type OptionsConsumerFunc func(*Consumer)

func (f OptionsConsumerFunc) apply(cons *Consumer) {
	f(cons)
}

// SetConsMsgDelayParams 开发者设置消费者端初始化时的参数
func SetConsMsgDelayParams(delayExchangeName string, enableDelayMsgPlugin bool) OptionsConsumer {
	return OptionsConsumerFunc(func(c *Consumer) {
		c.enableDelayMsgPlugin = enableDelayMsgPlugin
		c.config.ExchangeType = "x-delayed-message"
		c.config.ExchangeName = delayExchangeName
		// 延迟消息队列，交换机、消息全部设置为持久
		c.config.Durable = true
	})
}

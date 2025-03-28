package rocketmq

import "fmt"

type MsgChannel struct {
	Topic string // RocketMQ主题（对应RabbitMQ的Exchange）
	Tag   string // 消息标签（用于过滤，对应RabbitMQ的RoutingKey）
}

func (m *MsgChannel) Id() string {
	return fmt.Sprintf("t%s_tag%s_",
		m.Topic,
		m.Tag)
}

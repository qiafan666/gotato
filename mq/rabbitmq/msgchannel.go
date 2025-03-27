package rabbitmq

import "fmt"

type MsgChannel struct {
	Queue        string
	Exchange     string
	RoutingKey   string
	ExchangeType string
	DeclareQueue bool
}

func (m *MsgChannel) Id() string {
	return fmt.Sprintf("e%s_t%s_k%s_q%s", m.Exchange, m.ExchangeType, m.RoutingKey, m.Queue)
}

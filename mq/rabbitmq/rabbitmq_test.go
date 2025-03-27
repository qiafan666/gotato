package rabbitmq

import (
	"context"
	"fmt"
	"github.com/qiafan666/gotato/commons/gface"
	"testing"
)

type TestOrder struct {
}

func NewTestOrder() *TestOrder {
	return &TestOrder{}
}

func (t *TestOrder) Handle(msg string) error {
	fmt.Println("TestOrder Handle", msg)
	return nil
}

func TestConsumer(t *testing.T) {

	consumer, err := NewConsumer(context.Background(), "amqp://rabbitmq:D3oyMv9A6Vxc@10.0.0.222:5672/", gface.NewLogger("rabbitmq.consumer", nil))
	if err != nil {
		return
	}
	defer consumer.Close()

	testOrder := NewTestOrder()
	msgChannel := &MsgChannel{
		Queue:        "test_queue",
		Exchange:     "test_exchange",
		RoutingKey:   "test_routing_key",
		ExchangeType: "direct",
	}
	go consumer.Consume(context.Background(), msgChannel, testOrder)
	select {}
}

func TestProducer(t *testing.T) {
	producer, err := NewProducer(context.Background(), "", gface.NewLogger("rabbitmq.producer", nil))
	if err != nil {
		return
	}
	defer producer.Close()

	msgChannel := &MsgChannel{
		Queue:        "test_queue",
		Exchange:     "",
		RoutingKey:   "test_routing_key",
		ExchangeType: "direct",
	}
	err = producer.Publish(context.Background(), msgChannel, "msg")
	if err != nil {
		return
	}
}

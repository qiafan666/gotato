package rabbitmq

import (
	"context"
	"fmt"
	"github.com/qiafan666/gotato/commons/gapp/example/def"
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

	consumer, err := NewConsumer(context.Background(), "", gface.NewLogger("rabbitmq.consumer", def.ZapLog))
	if err != nil {
		return
	}
	defer consumer.Close()

	testOrder := NewTestOrder()
	msgChannel := &MsgChannel{
		Queue:        "test_queue",
		Exchange:     "",
		RoutingKey:   "test_routing_key",
		ExchangeType: "direct",
	}
	go consumer.Consume(context.Background(), msgChannel, testOrder)
}

func TestProducer(t *testing.T) {
	producer, err := NewProducer(context.Background(), "", gface.NewLogger("rabbitmq.producer", def.ZapLog))
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

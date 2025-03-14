package kafka

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

func (t *TestOrder) Handle(topic, message string) {
	fmt.Println(topic, message)
}

func TestProduce(t *testing.T) {
	producer := NewAsyncProducer([]string{"localhost:9092"}, "topic", gface.NewLogger("kafka.producer", def.ZapLog))
	go producer.Run(context.Background())

	err := producer.Push("test")
	if err != nil {
		return
	}
}

func TestConsumer(t *testing.T) {

	order := NewTestOrder()
	consumer := NewConsumer("consumer", []string{"topic"}, []string{"localhost:9092"}, order, gface.NewLogger("kafka.consumer", def.ZapLog))
	go StartConsumerGroup(context.Background(), "group", consumer)
}

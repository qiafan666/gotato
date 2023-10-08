package kafka

import (
	"fmt"
	"github.com/IBM/sarama"
)

type handler struct {
}

func (h handler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h handler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Println(string(msg.Value))
		session.MarkMessage(msg, "")
	}
	return nil
}

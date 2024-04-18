package kafka

import (
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"testing"
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

var (
	// brokerList 列表
	brokerList []string
	// topic 主题名称
	topic string
	// maxRetry 重试次数
	maxRetry int
)

func TestKafkaProducer(t *testing.T) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = maxRetry
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		log.Panic(err)
	}
	defer func() {
		if err = producer.Close(); err != nil {
			log.Panic(err)
		}
	}()
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder("Something Cool"),
	}
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", topic, partition, offset)
}

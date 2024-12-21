package kafka

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/qiafan666/gotato/commons/gerr"
	"google.golang.org/protobuf/proto"
)

// Producer 生产者
type Producer struct {
	addr     []string
	topic    string
	config   *sarama.Config
	producer sarama.SyncProducer
}

// NewKafkaProducer 创建生产者
func NewKafkaProducer(config *sarama.Config, addr []string, topic string) (*Producer, error) {
	producer, err := NewProducer(config, addr)
	if err != nil {
		return nil, err
	}
	return &Producer{
		addr:     addr,
		topic:    topic,
		config:   config,
		producer: producer,
	}, nil
}

func (p *Producer) SendMessage(ctx context.Context, key string, msg proto.Message) (int32, int64, error) {
	// Marshal the protobuf message
	bMsg, err := proto.Marshal(msg)
	if err != nil {
		return 0, 0, gerr.WrapMsg(err, "proto.Marshal error")
	}
	if len(bMsg) == 0 {
		return 0, 0, gerr.New("SendMessage msg is empty")
	}

	// 包装消息
	kMsg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(bMsg),
	}

	// 验证消息
	if kMsg.Key.Length() == 0 || kMsg.Value.Length() == 0 {
		return 0, 0, gerr.New("SendMessage key or value is empty")
	}

	// 添加消息头
	header, err := GetMQHeaderWithContext(ctx)
	if err != nil {
		return 0, 0, err
	}
	kMsg.Headers = header

	// 发送消息
	partition, offset, err := p.producer.SendMessage(kMsg)
	if err != nil {
		return 0, 0, gerr.WrapMsg(err, "p.producer.SendMessage error")
	}

	return partition, offset, nil
}

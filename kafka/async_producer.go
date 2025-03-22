package kafka

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/gface"
	"github.com/qiafan666/gotato/commons/gjson"
)

type AsyncProducer struct {
	addr     []string
	topic    string
	producer sarama.AsyncProducer
	logger   gface.Logger
}

func NewAsyncProducer(addr []string, topic string, logger gface.Logger) *AsyncProducer {
	p := &AsyncProducer{
		addr:   addr,
		topic:  topic,
		logger: logger,
	}

	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true
	cfg.Producer.Partitioner = sarama.NewHashPartitioner
	producer, err := sarama.NewAsyncProducer(p.addr, cfg)
	if err != nil {
		p.logger.ErrorF(nil, "Failed to create async producer: %v", err)
	}
	p.producer = producer

	return p
}

// Push kafka消息推送
// key: kafka消息的key,可为nil，分区采用轮训算法，用于将消息分区，如果要保证同类消息的顺序消费，则key必须设置，让消息进入同一个分区
// data: 待发送的消息内容，可以是任意类型，将会被序列化为json格式
// 如果设置了key，建议用consumerbatch，因为key可以保证同类消息的顺序消费，他就可以只提交最后一次的消息的offset
// 如果不设置key，建议用consumergroup，需要每次提交
func (p *AsyncProducer) Push(key sarama.StringEncoder, data interface{}) error {
	marshal, err := gjson.Marshal(data)
	if err != nil {
		return err
	}
	msg := &sarama.ProducerMessage{Topic: p.topic, Key: key, Value: sarama.StringEncoder(gcommon.Bytes2Str(marshal))}
	p.producer.Input() <- msg
	return nil
}

func (p *AsyncProducer) Run(ctx context.Context) {
	defer func() {
		if err := p.producer.Close(); err != nil {
			p.logger.ErrorF(nil, "Failed to close async producer: %v", err)
		}
	}()

	for {
		select {
		case err := <-p.producer.Errors():
			p.logger.WarnF(nil, "Failed to produce message topic: %v err: %v", p.topic, err)
		case msg := <-p.producer.Successes():
			msgBytes, err := msg.Value.Encode()
			if err != nil {
				p.logger.DebugF(nil, "Failed to encode message topic: %v err: %v", p.topic, err)
			} else {
				p.logger.DebugF(nil, "Successes to produce message topic: %v Value: %v", p.topic, gcommon.Bytes2Str(msgBytes))
			}
		case <-ctx.Done():
			p.logger.InfoF(nil, "AsyncProducer Quit.")
			return
		}
	}
}

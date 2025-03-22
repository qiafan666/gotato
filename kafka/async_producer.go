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
	producer, err := sarama.NewAsyncProducer(p.addr, cfg)
	if err != nil {
		p.logger.ErrorF(nil, "Failed to create async producer: %v", err)
	}
	p.producer = producer

	return p
}

func (p *AsyncProducer) Push(data interface{}) error {
	marshal, err := gjson.Marshal(data)
	if err != nil {
		return err
	}
	msg := &sarama.ProducerMessage{Topic: p.topic, Key: nil, Value: sarama.StringEncoder(gcommon.Bytes2Str(marshal))}
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

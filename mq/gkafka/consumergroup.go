package gkafka

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/gface"
)

type ConsumerGroupHandler struct {
	msgHandler IHandler
	topics     []string
	addr       []string
	logger     gface.ILogger
}

func (ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h ConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// 获取消息
	for message := range claim.Messages() {
		h.logger.DebugF(context.Background(), "ConsumeClaim: Message claimed: value = %s,topic = %s,partition=%d, offset=%+v", gcommon.Bytes2Str(message.Value), message.Topic, message.Partition, message.Offset)
		h.msgHandler.Handle(message.Topic, gcommon.Bytes2Str(message.Value))
		// 将消息标记为已使用
		sess.MarkMessage(message, "")
	}
	return nil
}

// NewConsumer 创建消费者
func NewConsumer(topics []string, addr []string, msgHandler IHandler, logger gface.ILogger) *ConsumerGroupHandler {
	handler := &ConsumerGroupHandler{
		msgHandler: msgHandler,
		topics:     topics,
		addr:       addr,
		logger:     logger,
	}
	return handler
}

func StartConsumerGroup(ctx context.Context, groupId string, handler *ConsumerGroupHandler) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = sarama.V2_0_0_0
	config.Consumer.Offsets.AutoCommit.Enable = true
	// consumer
	group, err := sarama.NewConsumerGroup(handler.addr, groupId, config)
	if err != nil {
		handler.logger.ErrorF(nil, "StartConsumerGroup: create consumer group error. err=%+v", err)
		return
	}
	defer group.Close()

	// 检查错误
	go func() {
		for err = range group.Errors() {
			handler.logger.ErrorF(nil, "StartConsumerGroup: consumer group error. err=%+v", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			err = group.Consume(ctx, handler.topics, handler)
			if err != nil {
				handler.logger.ErrorF(nil, "StartConsumerGroup: consume error. err=%+v", err)
			}
		}
	}
}

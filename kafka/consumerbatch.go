package kafka

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/qiafan666/gotato/commons/gbatcher"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/gface"
)

type ConsumerBatchHandler struct {
	msgHandler   IHandler
	consumerName string
	topics       []string
	addr         []string
	logger       gface.ILogger
	batcher      *gbatcher.Batcher[sarama.ConsumerMessage]
}

func (ConsumerBatchHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (ConsumerBatchHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h ConsumerBatchHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	//以最后一次消息的偏移量为准，提交偏移量，减少io操作
	//只有设置了key的情况下，保证同类数据进入同一分区才能按最后一次提交
	h.batcher.OnComplete = func(lastMessage *sarama.ConsumerMessage, totalCount int) {
		sess.MarkMessage(lastMessage, "")
		sess.Commit()
	}

	// 获取消息
	for message := range claim.Messages() {
		if len(message.Value) == 0 {
			continue
		}
		h.logger.DebugF(context.Background(), "ConsumeClaim: Message claimed: value = %s,topic = %s,partition=%d, offset=%+v", gcommon.Bytes2Str(message.Value), message.Topic, message.Partition, message.Offset)

		err := h.batcher.Put(context.Background(), message)
		if err != nil {
			h.logger.ErrorF(nil, "kafka consume batcher put message error. err=%+v", err)
		}
	}
	return nil
}

// NewConsumerBatch 创建一个批量处理的消费者，结合gbatcher使用
func NewConsumerBatch(
	consumerName string, topics []string, addr []string, msgHandler IHandler,
	logger gface.ILogger, batcher *gbatcher.Batcher[sarama.ConsumerMessage]) *ConsumerBatchHandler {
	handler := &ConsumerBatchHandler{
		msgHandler:   msgHandler,
		consumerName: consumerName,
		topics:       topics,
		addr:         addr,
		logger:       logger,
		batcher:      batcher,
	}
	// 设置key分组方法，这里使用消息的key作为分组的依据，同类消息顺序消费
	handler.batcher.Key = func(consumerMessage *sarama.ConsumerMessage) string {
		return string(consumerMessage.Key)
	}

	// 设置处理方法
	handler.batcher.Do = func(ctx context.Context, _ int, val *gbatcher.Msg[sarama.ConsumerMessage]) {
		for _, msg := range val.Val() {
			handler.msgHandler.Handle(msg.Topic, gcommon.Bytes2Str(msg.Value))
		}
	}
	return handler
}

func StartConsumerBatch(ctx context.Context, groupId string, handler *ConsumerBatchHandler) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = sarama.V0_11_0_2
	// consumer
	group, err := sarama.NewConsumerGroup(handler.addr, groupId, config)
	if err != nil {
		handler.logger.ErrorF(ctx, "StartConsumerBatch: create consumer group error. err=%+v", err)
		return
	}

	// 检查错误
	go func() {
		for err = range group.Errors() {
			handler.logger.ErrorF(ctx, "StartConsumerBatch: consumer group error. err=%+v", err)
		}
	}()

	for {
		err = group.Consume(ctx, handler.topics, handler)
		if err != nil {
			handler.logger.ErrorF(ctx, "StartConsumerBatch: consume error. err=%+v", err)
		}
		if ctx.Err() != nil {
			return
		}
	}
}

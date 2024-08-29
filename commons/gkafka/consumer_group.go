package gkafka

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"github.com/qiafan666/gotato/commons/glog"
)

type MConsumerGroup struct {
	sarama.ConsumerGroup
	groupID string
	topics  []string
}

// NewMConsumerGroup 创建一个新的消费者组
func NewMConsumerGroup(conf *Config, groupID string, topics []string, autoCommitEnable bool) (*MConsumerGroup, error) {
	config, err := BuildConsumerGroupConfig(conf, sarama.OffsetNewest, autoCommitEnable)
	if err != nil {
		return nil, err
	}
	group, err := NewConsumerGroup(config, conf.Addr, groupID)
	if err != nil {
		return nil, err
	}
	return &MConsumerGroup{
		ConsumerGroup: group,
		groupID:       groupID,
		topics:        topics,
	}, nil
}

// RegisterHandleAndConsumer 注册处理函数和消费者
func (mc *MConsumerGroup) RegisterHandleAndConsumer(ctx context.Context, handler sarama.ConsumerGroupHandler) {
	for {
		err := mc.ConsumerGroup.Consume(ctx, mc.topics, handler)
		if errors.Is(err, sarama.ErrClosedConsumerGroup) {
			return
		}
		if errors.Is(err, context.Canceled) {
			return
		}
		if err != nil {
			glog.Slog.ErrorKVs(ctx, "consume err", err, "topic", mc.topics, "groupID", mc.groupID)
		}
	}
}

// Close 关闭消费者组
func (mc *MConsumerGroup) Close() error {
	return mc.ConsumerGroup.Close()
}

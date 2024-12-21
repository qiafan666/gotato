package kafka

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/qiafan666/gotato/commons/gcast"
	"github.com/qiafan666/gotato/commons/gerr"
)

func Check(ctx context.Context, conf *Config, topics []string) error {
	kfk, err := BuildConsumerGroupConfig(conf, sarama.OffsetNewest, false)
	if err != nil {
		return err
	}
	cli, err := sarama.NewClient(conf.Addr, kfk)
	if err != nil {
		return gerr.WrapMsg(err, "Failed to create kafka client", "config", gcast.ToString(conf))
	}
	defer cli.Close()

	existingTopics, err := cli.Topics()
	if err != nil {
		return gerr.WrapMsg(err, "Failed to list topics")
	}

	existingTopicsMap := make(map[string]bool)
	for _, t := range existingTopics {
		existingTopicsMap[t] = true
	}

	for _, topic := range topics {
		if !existingTopicsMap[topic] {
			return gerr.New(fmt.Sprintf("Topic %s not exist", topic))
		}
	}
	return nil
}

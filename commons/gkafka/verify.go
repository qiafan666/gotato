package gkafka

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/pkg/errors"
	"github.com/qiafan666/gotato/commons/gcast"
)

func Check(ctx context.Context, conf *Config, topics []string) error {
	kfk, err := BuildConsumerGroupConfig(conf, sarama.OffsetNewest, false)
	if err != nil {
		return err
	}
	cli, err := sarama.NewClient(conf.Addr, kfk)
	if err != nil {
		return errors.Wrapf(err, "NewClient failed,config: %s", gcast.ToString(conf))
	}
	defer cli.Close()

	existingTopics, err := cli.Topics()
	if err != nil {
		return errors.Wrap(err, "Failed to list topics")
	}

	existingTopicsMap := make(map[string]bool)
	for _, t := range existingTopics {
		existingTopicsMap[t] = true
	}

	for _, topic := range topics {
		if !existingTopicsMap[topic] {
			return errors.New(fmt.Sprintf("Topic %s not exist", topic))
		}
	}
	return nil
}

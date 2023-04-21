package utils

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	slog "github.com/qiafan666/quickweb/commons/log"
)

func KafkaSend(topic string, key string, data []byte, host string, port uint) error {
	saramaConf := sarama.NewConfig()
	saramaConf.Producer.RequiredAcks = sarama.WaitForAll          //赋值为-1：这意味着producer在follower副本确认接收到数据后才算一次发送完成。
	saramaConf.Producer.Partitioner = sarama.NewRandomPartitioner //写到随机分区中，默认设置8个分区
	saramaConf.Producer.Return.Successes = true
	msg := &sarama.ProducerMessage{}
	msg.Topic = topic
	msg.Value = sarama.StringEncoder(data)
	msg.Key = sarama.StringEncoder(key)
	client, err := sarama.NewSyncProducer([]string{fmt.Sprintf("%s:%d", host, port)}, saramaConf)
	if err != nil {
		slog.Slog.ErrorF(context.Background(), "kafka connection err["+err.Error()+"]")
		return err
	}
	defer client.Close()
	if _, _, err := client.SendMessage(msg); err != nil {
		slog.Slog.ErrorF(context.Background(), "kafka send failed["+err.Error()+"]")
		return err
	}
	return nil
}

package kafka

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/qiafan666/gotato/commons/glog"
	"github.com/qiafan666/gotato/gconfig"
	"strconv"
	"sync"
)

func Receiver(ctx context.Context, topic string, callBackChan chan []byte) {

	go func() {
		wg := sync.WaitGroup{}
		// 根据给定的代理地址和配置创建一个消费者
		consumer, err := sarama.NewConsumer([]string{gconfig.Configs.Kafka.Host + ":" + strconv.Itoa(gconfig.Configs.Kafka.Port)}, nil)

		if err != nil {
			return
		}
		//Partitions(topic):该方法返回了该topic的所有分区id
		partitionList, err := consumer.Partitions(topic)
		if err != nil {
			return
		}
		for partition := range partitionList {
			//ConsumePartition方法根据主题，分区和给定的偏移量创建创建了相应的分区消费者
			//如果该分区消费者已经消费了该信息将会返回error
			//sarama.OffsetNewest:表明了为最新消息
			pc, err := consumer.ConsumePartition(topic, int32(partition), sarama.OffsetNewest)
			if err != nil {
				return
			}
			defer pc.AsyncClose()
			wg.Add(1)
			go func(sarama.PartitionConsumer) {
				defer wg.Done()
				//Messages()该方法返回一个消费消息类型的只读通道，由代理产生
				select {
				case msg, ok := <-pc.Messages():
					if ok {
						callBackChan <- msg.Value
					} else {
						return
					}
				case <-ctx.Done():
					return

				}
			}(pc)
		}
		wg.Wait()
		consumer.Close().Error()
	}()
}

func GroupReceiver(ctx context.Context, brokers []string, group string, topics []string, handler sarama.ConsumerGroupHandler) error {

	wg := sync.WaitGroup{}
	wg.Add(1)
	configK := sarama.NewConfig()
	configK.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategyRange}
	configK.Consumer.Offsets.Initial = sarama.OffsetOldest
	glog.Slog.InfoF(ctx, "Error creating consumer group client connecting brokers %+v", brokers)
	consumer, err := sarama.NewConsumerGroup(brokers, group, configK)
	if err != nil {
		glog.Slog.InfoF(ctx, "Error creating consumer group client: %+v", err)
		return err
	}
	glog.Slog.InfoF(ctx, "Error creating consumer group client connected")
	defer func() {
		if err = consumer.Close(); err != nil {
			glog.Slog.InfoF(ctx, "Error creating consumer group client: %+v", err)
		}
	}()
	go func() {
		defer wg.Done()
		for {
			if err := consumer.Consume(ctx, topics, handler); err != nil {
				glog.Slog.InfoF(ctx, "Error creating consumer group client: %+v", err)
			}

			if ctx.Done() != nil {
				return
			}
		}
	}()

	wg.Wait()
	return nil
}

func Send(topic string, key string, data []byte, host string, port uint) error {
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
		glog.Slog.ErrorF(context.Background(), "kafka connection err["+err.Error()+"]")
		return err
	}
	defer func(client sarama.SyncProducer) {
		err = client.Close()
		if err != nil {
			glog.Slog.ErrorF(context.Background(), "kafka close err["+err.Error()+"]")
		}
	}(client)
	if _, _, err = client.SendMessage(msg); err != nil {
		glog.Slog.ErrorF(context.Background(), "kafka send failed["+err.Error()+"]")
		return err
	}
	return nil
}

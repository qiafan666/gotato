package kafka

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/qiafan666/quickweb/config"
	"strconv"
	"sync"
)

type Kafka struct {
}

func (slf *Kafka) KafkaReceiver(ctx context.Context, topic string, callBackChan chan []byte) {

	go func() {
		wg := sync.WaitGroup{}
		// 根据给定的代理地址和配置创建一个消费者
		consumer, err := sarama.NewConsumer([]string{config.Configs.Kafka.Host + ":" + strconv.Itoa(int(config.Configs.Kafka.Port))}, nil)

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

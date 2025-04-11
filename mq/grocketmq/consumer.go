package grocketmq

import (
	"context"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/gface"
	"github.com/sirupsen/logrus"
	"log"
	"sync"
	"time"
)

const (
	PUSH                         = false
	PULL                         = true
	RefreshPersistOffsetDuration = 5 * time.Second
)

type Consumer struct {
	pushConsumer      rocketmq.PushConsumer
	pullConsumer      rocketmq.PullConsumer
	msgChannelDeclare sync.Map
	logger            gface.ILogger
	options           []consumer.Option
	mode              bool //push or pull false:push true:pull
}

func NewConsumer(ctx context.Context, logger gface.ILogger, mode bool, options ...consumer.Option) (*Consumer, error) {
	c := &Consumer{
		logger:  logger,
		options: options,
		mode:    mode,
	}

	err := c.init()
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		c.Close()
	}()

	return c, nil
}

func (c *Consumer) init() error {

	if c.mode && c.pullConsumer == nil {
		consume, err := rocketmq.NewPullConsumer(c.options...)
		if err != nil {
			c.logger.ErrorF(nil, "create fail. err=%+v", err)
			return err
		}
		c.pullConsumer = consume
	}

	if !c.mode && c.pushConsumer == nil {
		consume, err := rocketmq.NewPushConsumer(c.options...)
		if err != nil {
			c.logger.ErrorF(nil, "create fail. err=%+v", err)
			return err
		}
		c.pushConsumer = consume
	}

	return nil
}

func (c *Consumer) Close() {
	if c.mode && c.pullConsumer != nil {
		_ = c.pullConsumer.Shutdown()
	}

	if !c.mode && c.pushConsumer != nil {
		_ = c.pushConsumer.Shutdown()
	}
}

func (c *Consumer) Consume(ctx context.Context, msgChannel *MsgChannel, handler IHandler) {
	if ctx.Err() != nil {
		return
	}

	err := c.init()
	if err != nil {
		c.logger.ErrorF(nil, "init fail. retry after 5 seconds, err=%+v", err)
		<-time.After(5 * time.Second)
		go c.Consume(ctx, msgChannel, handler)
		return
	}

	err = c.subscribe(ctx, msgChannel, handler)
	if err != nil {
		c.logger.ErrorF(nil, "subscribe fail. retry after 5 seconds, err=%+v", err)
		<-time.After(5 * time.Second)
		go c.Consume(ctx, msgChannel, handler)
		return
	}

	// 启动消费者
	if !c.mode {
		// Push 模式
		if err = c.pushConsumer.Start(); err != nil {
			c.logger.ErrorF(nil, "pushConsumer start fail. err: %+v", err)
			time.Sleep(5 * time.Second)
			go c.Consume(ctx, msgChannel, handler)
			return
		}
	} else {
		if err = c.pullConsumer.Start(); err != nil {
			c.logger.ErrorF(nil, "pullConsumer start fail. err: %+v", err)
			time.Sleep(5 * time.Second)
			go c.Consume(ctx, msgChannel, handler)
		}
	}

	<-ctx.Done()
}

func (c *Consumer) subscribe(ctx context.Context, msgChannel *MsgChannel, handler IHandler) error {
	id := msgChannel.Id()
	if _, ok := c.msgChannelDeclare.Load(id); ok {
		return nil
	}

	selector := consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: msgChannel.Tag,
	}

	if c.mode {
		err := c.pullConsumer.Subscribe(msgChannel.Topic, selector)
		if err != nil {
			c.logger.ErrorF(nil, "subscribe fail. topic=%s, tag=%s, err=%+v", msgChannel.Topic, msgChannel.Tag, err)
			return err
		}

		timer := time.NewTimer(RefreshPersistOffsetDuration)
		go func() {
			for ; true; <-timer.C {
				err = c.pullConsumer.PersistOffset(ctx, msgChannel.Topic)
				if err != nil {
					log.Printf("[pullConsumer.PersistOffset] err=%v", err)
				}
				timer.Reset(RefreshPersistOffsetDuration)
			}
		}()

		go c.pullMessages(ctx, handler)
	} else {
		err := c.pushConsumer.Subscribe(msgChannel.Topic, selector,
			func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
				for _, msg := range msgs {
					c.logger.DebugF(nil, "receive msg. topic=%s,tag=%s, msgId=%s, body=%s",
						msg.Topic, msg.GetTags(), msg.MsgId, gcommon.Bytes2Str(msg.Body))

					err := handler.Handle(msg.GetTags(), gcommon.Bytes2Str(msg.Body))
					if err != nil {
						c.logger.ErrorF(nil, "handle msg fail. msgId=%s, err=%+v", msg.MsgId, err)
						return consumer.ConsumeRetryLater, nil
					}
				}
				return consumer.ConsumeSuccess, nil
			})

		if err != nil {
			c.logger.ErrorF(nil, "subscribe fail. topic=%s, tag=%s, err=%+v", msgChannel.Topic, msgChannel.Tag, err)
			return err
		}
	}

	c.msgChannelDeclare.Store(id, struct{}{}) // 订阅成功后才存入
	return nil
}

// Pull 模式消息轮询
func (c *Consumer) pullMessages(ctx context.Context, handler IHandler) {

	batchSize := 10 // 初始拉取数量

	for {
		select {
		case <-ctx.Done():
			return
		default:
			startTime := time.Now()

			resp, err := c.pullConsumer.Pull(ctx, batchSize)
			if err != nil {
				c.logger.ErrorF(nil, "pullMessages pull fail. err: %+v", err)
				time.Sleep(5 * time.Second) // 避免频繁请求
				continue
			}

			if len(resp.GetMessages()) == 0 {
				time.Sleep(time.Second) // 没有消息，稍作等待
				continue
			}

			switch resp.Status {
			case primitive.PullFound:
				logrus.Debugf("[pull message successfully] MinOffset:%d, MaxOffset:%d, nextOffset: %d, len:%d\n", resp.MinOffset, resp.MaxOffset, resp.NextBeginOffset, len(resp.GetMessages()))
				var queue *primitive.MessageQueue
				if len(resp.GetMessages()) <= 0 {
					time.Sleep(time.Second)
					continue
				}
				for _, msg := range resp.GetMessages() {
					queue = msg.Queue
					if err = handler.Handle(msg.GetTags(), gcommon.Bytes2Str(msg.Body)); err != nil {
						c.logger.ErrorF(nil, "handle msg fail.err=%+v", err)
					}
				}

				err = c.pullConsumer.UpdateOffset(queue, resp.NextBeginOffset)
				if err != nil {
					c.logger.ErrorF(nil, "updates offset fail. err: %+v", err)
					continue
				}

			case primitive.PullNoNewMsg, primitive.PullNoMsgMatched:
				c.logger.ErrorF(nil, "pullMessages no new message. status: %d", resp.Status)
				time.Sleep(time.Second)
				return
			case primitive.PullBrokerTimeout:
				c.logger.ErrorF(nil, "pullBrokerTimeout")

				time.Sleep(time.Second)
				return
			case primitive.PullOffsetIllegal:
				c.logger.ErrorF(nil, "pull offset illegal")
				return
			default:
				c.logger.ErrorF(nil, "pullMessages unknown status: %d", resp.Status)
			}

			elapsed := time.Since(startTime)
			if elapsed < 500*time.Millisecond && batchSize < 100 {
				batchSize += 5
			} else if elapsed > time.Second && batchSize > 10 {
				batchSize -= 5
				if batchSize < 10 {
					batchSize = 10
				}
			}
		}
	}
}

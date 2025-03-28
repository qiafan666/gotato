package rocketmq

import (
	"context"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/gface"
	"sync"
	"time"
)

type Consumer struct {
	consumer          rocketmq.PushConsumer
	msgChannelDeclare sync.Map
	logger            gface.ILogger
	options           []consumer.Option
}

func NewConsumer(ctx context.Context, logger gface.ILogger, options ...consumer.Option) (*Consumer, error) {
	c := &Consumer{
		logger:  logger,
		options: options,
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
	if c.consumer == nil {
		consume, err := rocketmq.NewPushConsumer(c.options...)
		if err != nil {
			c.logger.ErrorF(nil, "create fail. err=%+v", err)
			return err
		}

		c.consumer = consume
	}
	return nil
}

func (c *Consumer) Close() {
	if c.consumer != nil {
		_ = c.consumer.Shutdown()
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

	err = c.subscribe(msgChannel, handler)
	if err != nil {
		c.logger.ErrorF(nil, "subscribe fail. retry after 5 seconds, err=%+v", err)
		<-time.After(5 * time.Second)
		go c.Consume(ctx, msgChannel, handler)
		return
	}

	err = c.consumer.Start()
	if err != nil {
		c.logger.ErrorF(nil, "start fail. retry after 5 seconds, err=%+v", err)
		<-time.After(5 * time.Second)
		go c.Consume(ctx, msgChannel, handler)
		return
	}

	<-ctx.Done()
}

func (c *Consumer) subscribe(msgChannel *MsgChannel, handler IHandler) error {
	id := msgChannel.Id()
	if _, ok := c.msgChannelDeclare.Load(id); ok {
		return nil
	}

	selector := consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: msgChannel.Tag,
	}

	err := c.consumer.Subscribe(msgChannel.Topic, selector,
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

	c.msgChannelDeclare.Store(id, struct{}{}) // 订阅成功后才存入
	return nil
}

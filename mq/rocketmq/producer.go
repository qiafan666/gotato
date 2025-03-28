package rocketmq

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/gface"
	"github.com/qiafan666/gotato/commons/gson"
)

type Producer struct {
	producer     rocketmq.Producer
	topicDeclare sync.Map
	logger       gface.ILogger
	options      []producer.Option
}

func NewProducer(ctx context.Context, logger gface.ILogger, options ...producer.Option) (*Producer, error) {
	p := &Producer{
		logger:  logger,
		options: options,
	}

	err := p.init()
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		p.Close()
	}()

	return p, nil
}

func (p *Producer) init() error {
	if p.producer == nil {
		produce, err := rocketmq.NewProducer(
			p.options...,
		)
		if err != nil {
			p.logger.ErrorF(nil, "RocketMQ.NewProducer create fail.err=%+v", err)
			return err
		}

		err = produce.Start()
		if err != nil {
			p.logger.ErrorF(nil, "RocketMQ.Producer start fail. err=%+v", err)
			return err
		}

		p.producer = produce
	}
	return nil
}

func (p *Producer) Close() {
	if p.producer != nil {
		_ = p.producer.Shutdown()
	}
}

func (p *Producer) Publish(ctx context.Context, msgChannel *MsgChannel, msg interface{}) error {
	marshal, err := gson.Marshal(msg)
	if err != nil {
		p.logger.ErrorF(nil, "Marshal fail. topic=%s, msg=%+v, err=%+v",
			msgChannel.Topic, msg, err)
		return err
	}
	now := time.Now()

	if err = p.init(); err != nil {
		p.logger.ErrorF(nil, "Init fail. topic=%s, msg=%+v, err=%+v",
			msgChannel.Topic, gcommon.Bytes2Str(marshal), err)
		return err
	}

	// 构建RocketMQ消息
	rmqMsg := &primitive.Message{
		Topic: msgChannel.Topic,
		Body:  marshal,
	}

	// 设置消息属性
	rmqMsg.WithKeys([]string{strconv.FormatInt(now.UnixNano(), 10)})
	rmqMsg.WithTag(msgChannel.Tag)
	// 发送消息
	res, err := p.producer.SendSync(ctx, rmqMsg)
	if err != nil {
		p.logger.ErrorF(nil, "SendSync fail. topic=%s, tag=%s, msg=%+v, err=%+v",
			msgChannel.Topic, msgChannel.Tag, gcommon.Bytes2Str(marshal), err)
		return err
	}

	p.logger.DebugF(nil, "SendSync success. topic=%s, tag=%s, msgId=%s, queueOffset=%d,msg=%+v",
		msgChannel.Topic, msgChannel.Tag, res.MsgID, res.QueueOffset, gcommon.Bytes2Str(marshal))
	return nil
}

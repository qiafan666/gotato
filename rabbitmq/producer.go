package rabbitmq

import (
	"context"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/gface"
	"github.com/qiafan666/gotato/commons/gson"
	amqp "github.com/rabbitmq/amqp091-go"
	"strconv"
	"sync"
	"time"
)

type Producer struct {
	conn              *amqp.Connection
	channel           *amqp.Channel
	msgChannelDeclare sync.Map

	url    string
	logger gface.Logger
}

func NewProducer(ctx context.Context, url string, logger gface.Logger) (*Producer, error) {
	producer := &Producer{
		url:    url,
		logger: logger,
	}

	err := producer.ini()
	if err != nil {
		return nil, err
	}

	go func() {
		select {
		case <-ctx.Done():
			producer.Close()
		}
	}()

	return producer, nil
}

func (p *Producer) ini() error {
	if p.conn == nil || p.conn.IsClosed() {
		conn, err := amqp.Dial(p.url)
		if err != nil {
			p.logger.ErrorF(nil, "RabbitMQ. NewProducer dial fail. url=%+v, err=%+v", p.url, err)
			return err
		}

		p.conn = conn
	}

	if p.channel == nil || p.channel.IsClosed() {
		channel, err := p.conn.Channel()
		if err != nil {
			p.logger.ErrorF(nil, "RabbitMQ. NewProducer open channel fail. url=%+v, err=%+v", p.url, err)
			return err
		}
		p.channel = channel
	}

	return nil
}

func (p *Producer) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}

func (p *Producer) Publish(ctx context.Context, msgChannel *MsgChannel, msg interface{}) error {
	marshal, _ := gson.Marshal(msg)
	now := time.Now()

	if err := p.ini(); err != nil {
		p.logger.ErrorF(nil, "RabbitMQ.Producer: Init fail. msgChannel=%+v, msg=%+v, err=%+v", msgChannel, gcommon.Bytes2Str(marshal), err)
		return err
	}
	err := p.declare(msgChannel)
	if err != nil {
		return err
	}
	err = p.channel.PublishWithContext(ctx, msgChannel.Exchange, msgChannel.RoutingKey, true, false, amqp.Publishing{
		MessageId: strconv.FormatInt(now.UnixNano(), 10),
		Timestamp: now,
		Body:      marshal,
	})
	if err != nil {
		p.logger.ErrorF(nil, "RabbitMQ.Producer: PublishWithContext fail. msgChannel=%+v, msg=%+v, err=%+v", msgChannel, gcommon.Bytes2Str(marshal), err)
		return err
	}
	p.logger.DebugF(nil, "RabbitMQ.Producer: PublishWithContext success. msgChannel=%+v, msg=%+v", msgChannel, gcommon.Bytes2Str(marshal))
	return nil
}

func (p *Producer) declare(msgChannel *MsgChannel) error {
	id := msgChannel.Id()
	_, ok := p.msgChannelDeclare.Load(id)
	if ok {
		return nil
	}
	err := p.channel.ExchangeDeclare(msgChannel.Exchange, msgChannel.ExchangeType, true, false, false, false, nil)
	if err != nil {
		p.logger.ErrorF(nil, "RabbitMQ.Producer: Exchange Declare fail. msgChannel=%+v, err=%+v", msgChannel, err)
		return err
	}
	p.msgChannelDeclare.Store(id, id)

	if msgChannel.DeclareQueue && msgChannel.Queue != "" {
		queue, err := p.channel.QueueDeclare(msgChannel.Queue, true, false, false, false, nil)
		if err != nil {
			p.logger.ErrorF(nil, "RabbitMQ.Producer: Queue Declare fail. msgChannel=%+v, err=%+v", msgChannel, err)
			return err
		}

		err = p.channel.QueueBind(queue.Name, msgChannel.RoutingKey, msgChannel.Exchange, false, nil)
		if err != nil {
			p.logger.ErrorF(nil, "RabbitMQ.Producer: Queue Bind fail. msgChannel=%+v, queue=%+v, err=%+v", msgChannel, queue.Name, err)
			return err
		}
	}

	return nil
}

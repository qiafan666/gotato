package rabbitmq

import (
	"context"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/gface"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
	"time"
)

type Consumer struct {
	conn              *amqp.Connection
	msgChannelDeclare sync.Map

	options map[string]any

	url string

	logger gface.ILogger
}

func NewConsumer(ctx context.Context, url string, logger gface.ILogger) (*Consumer, error) {
	c := &Consumer{
		url:     url,
		options: map[string]any{},
		logger:  logger,
	}
	err := c.ini()
	if err != nil {
		return nil, err
	}

	go func() {
		select {
		case <-ctx.Done():
			c.Close()
		}
	}()

	return c, nil
}

func (c *Consumer) ini() error {
	if c.conn == nil || c.conn.IsClosed() {
		conn, err := amqp.Dial(c.url)
		if err != nil {
			c.logger.ErrorF(nil, "RabbitMQ. NewConsumer dial fail. url=%+v, err=%+v", c.url, err)
			return err
		}

		c.conn = conn
	}

	return nil
}

func (c *Consumer) Close() {
	c.conn.Close()
}

func (c *Consumer) Consume(ctx context.Context, msgChannel *MsgChannel, handler IHandler) {
	if ctx.Err() != nil {
		return
	}
	err := c.ini()
	if err != nil {
		c.logger.ErrorF(nil, "RabbitMQ.Consumer: Consumer init fail. retry after 5 seconds, err=%+v", err)
		<-time.After(5 * time.Second)
		go c.Consume(ctx, msgChannel, handler)
		return
	}

	channel, err := c.conn.Channel()
	if err != nil {
		c.logger.ErrorF(nil, "RabbitMQ.Consumer: Consumer open channel fail. retry after 5 seconds, err=%+v", err)
		<-time.After(5 * time.Second)
		go c.Consume(ctx, msgChannel, handler)
		return
	}
	defer channel.Close()

	queue, err := c.declare(channel, msgChannel)
	if err != nil {
		c.logger.ErrorF(nil, "RabbitMQ.Consumer: Consumer declare queue fail. retry after 5 seconds, err=%+v", err)
		<-time.After(5 * time.Second)
		go c.Consume(ctx, msgChannel, handler)
		return
	}
	m, err := channel.Consume(queue.Name, msgChannel.Id(), false, false, false, false, nil)
	if err != nil {
		c.logger.ErrorF(nil, "RabbitMQ.Consumer: Consumer consume fail. retry after 5 seconds, err=%+v", err)
		<-time.After(5 * time.Second)
		go c.Consume(ctx, msgChannel, handler)
		return
	}
	c.logger.InfoF(nil, "RabbitMQ.Consumer: Consumer start. msgChannel=%+v, queue=%+v", msgChannel, queue.Name)

	connClose := make(chan *amqp.Error, 1)
	channelClose := make(chan *amqp.Error, 1)
	c.conn.NotifyClose(connClose)
	channel.NotifyClose(channelClose)

	for {
		select {
		case <-ctx.Done():
			return
		case amqpErr := <-connClose:
			if amqpErr == nil {
				return
			}
			c.logger.ErrorF(nil, "RabbitMQ.Consumer: Consumer connection close. reconnecting, err=%+v", amqpErr)
			c.msgChannelDeclare.Delete(msgChannel.Id())
			go c.Consume(ctx, msgChannel, handler)
			return
		case amqpErr := <-channelClose:
			if amqpErr == nil {
				return
			}
			c.logger.ErrorF(nil, "RabbitMQ.Consumer: Consumer channel close. reconnecting, err=%+v", amqpErr)
			c.msgChannelDeclare.Delete(msgChannel.Id())
			go c.Consume(ctx, msgChannel, handler)
			return
		case msg := <-m:
			c.logger.DebugF(nil, "RabbitMQ.Consumer: Consumer receive msg. msgChannel=%+v, queue=%+v, msgId=%+v, msgBody=%+v", msgChannel, queue.Name, msg.MessageId, gcommon.Bytes2Str(msg.Body))
			err = handler.Handle(gcommon.Bytes2Str(msg.Body))
			if err != nil {
				msg.Nack(false, true)
				continue
			}
			msg.Ack(false)
		}
	}

}

func (c *Consumer) declare(channel *amqp.Channel, msgChannel *MsgChannel) (*amqp.Queue, error) {
	id := msgChannel.Id()
	q, ok := c.msgChannelDeclare.Load(id)
	if ok {
		return q.(*amqp.Queue), nil
	}
	err := channel.ExchangeDeclare(msgChannel.Exchange, msgChannel.ExchangeType, true, false, false, false, nil)
	if err != nil {
		c.logger.ErrorF(nil, "RabbitMQ.Consumer: Exchange Declare fail. msgChannel=%+v, err=%+v", msgChannel, err)
		return nil, err
	}

	autoDelete := getOption[bool](c, "autoDelete", false)
	queue, err := channel.QueueDeclare(msgChannel.Queue, true, autoDelete, false, false, nil)
	if err != nil {
		c.logger.ErrorF(nil, "RabbitMQ.Consumer: Queue Declare fail. msgChannel=%+v, err=%+v", msgChannel, err)
		return nil, err
	}

	err = channel.QueueBind(queue.Name, msgChannel.RoutingKey, msgChannel.Exchange, false, nil)
	if err != nil {
		c.logger.ErrorF(nil, "RabbitMQ.Consumer: Queue Bind fail. msgChannel=%+v, queue=%+v, err=%+v", msgChannel, queue.Name, err)
		return nil, err
	}

	c.msgChannelDeclare.Store(id, &queue)
	return &queue, nil
}

func (c *Consumer) SetOption(key string, v any) {
	c.options[key] = v
}

func getOption[K any](c *Consumer, key string, defaultValue K) K {
	v, ok := c.options[key]
	if !ok {
		return defaultValue
	}

	r, ok := v.(K)
	if !ok {
		return defaultValue
	}
	return r
}

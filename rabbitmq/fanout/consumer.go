package fanout

import (
	"context"
	"fmt"
	"github.com/qiafan666/gotato/commons/log"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

func CreateConsumer(config ConsumerConfig, options ...OptionsConsumer) (*Consumer, error) {

	if config.Ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	if config.ChanNumber <= 0 {
		return nil, fmt.Errorf("channel number is less than 1")
	}

	conn, err := amqp.Dial(config.Addr)
	if err != nil {
		log.Slog.ErrorF(config.Ctx, "rabbitmq consumer connect error:%v", err)
		return nil, err
	}

	cons := &Consumer{
		ctx:                 config.Ctx,
		connect:             conn,
		connErr:             conn.NotifyClose(make(chan *amqp.Error, 1)),
		receivedMsgBlocking: make(chan struct{}),
		status:              1,
	}
	// rabbitmq 如果启动了延迟消息队列模式。继续初始化一些参数
	for _, val := range options {
		val.apply(cons)
	}
	return cons, nil
}

// Consumer 定义一个消息队列结构体：PublishSubscribe 模型
type Consumer struct {
	ctx                  context.Context
	config               ConsumerConfig
	connect              *amqp.Connection
	connErr              chan *amqp.Error
	callbackForReceived  func(receivedData []byte) //   断线重连，结构体内部使用
	callbackOffLine      func(err *amqp.Error)     //   断线重连，结构体内部使用
	enableDelayMsgPlugin bool                      // 是否使用延迟队列模式
	receivedMsgBlocking  chan struct{}             // 接受消息时用于阻塞消息处理函数
	status               byte                      // 客户端状态：1=正常；0=异常
	ch                   *amqp.Channel             // 通道
}

func (c *Consumer) Received(callbackFunDealMsg func(receivedData []byte)) {
	defer c.close()

	// 将回调函数地址赋值给结构体变量，用于掉线重连使用
	c.callbackForReceived = callbackFunDealMsg

	// 使用多个goroutine监听消息
	for i := 1; i <= c.config.ChanNumber; i++ {
		go c.consumeMessages(i, callbackFunDealMsg)
	}

	// 阻塞消息处理函数，防止客户端掉线后，消息处理函数继续执行
	if _, isOk := <-c.receivedMsgBlocking; isOk {
		c.status = 0
		close(c.receivedMsgBlocking)
	}
}

func (c *Consumer) consumeMessages(chanNo int, callbackFunDealMsg func(receivedData []byte)) {
	ch, err := c.connect.Channel()
	if err != nil {
		log.Slog.ErrorF(c.ctx, "rabbitmq consumer connect Channel error:%v, chanNo:%d", err, chanNo)
		return
	}
	defer ch.Close()

	c.ch = ch

	if err = c.setupExchangeAndQueue(ch); err != nil {
		log.Slog.ErrorF(c.ctx, "rabbitmq setup error: %s, chanNo:%d", err, chanNo)
		return
	}

	msgs, err := c.consumeMessagesFromChannel(ch)
	if err != nil {
		log.Slog.ErrorF(c.ctx, "rabbitmq consume error: %s, chanNo:%d", err, chanNo)
		return
	}

	c.processMessages(msgs, callbackFunDealMsg)
}

func (c *Consumer) setupExchangeAndQueue(ch *amqp.Channel) error {
	// 声明exchange交换机
	err := ch.ExchangeDeclare(
		c.config.ExchangeName, //exchange name
		c.config.ExchangeType, //exchange kind
		c.config.Durable,      //数据是否持久化
		!c.config.Durable,     //所有连接断开时，交换机是否删除
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// 声明队列
	queue, err := ch.QueueDeclare(
		c.config.QueueName,
		c.config.Durable,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	//队列绑定
	err = ch.QueueBind(
		queue.Name,
		"", //direct key，  fanout 模式设置为 空 即可
		c.config.ExchangeName,
		false,
		nil,
	)
	return err
}

// SetQos 设置质量保证
func (c *Consumer) SetQos(prefetchCount int, prefetchSize int, global bool) error {
	err := c.ch.Qos(
		prefetchCount, // 预取计数
		prefetchSize,  // 预取大小
		global,        // 全局应用
	)
	if err != nil {
		log.Slog.ErrorF(c.ctx, "设置Qos失败: %s", err.Error())
	}
	return err
}
func (c *Consumer) consumeMessagesFromChannel(ch *amqp.Channel) (<-chan amqp.Delivery, error) {
	// 消费消息
	return ch.ConsumeWithContext(
		c.ctx,
		c.config.QueueName, //queue name
		"",                 //consumer tag
		true,               //auto ack
		false,              //exclusive
		false,              //no local
		false,              //no wait
		nil,                //arguments
	)
}

func (c *Consumer) processMessages(msgs <-chan amqp.Delivery, callbackFunDealMsg func(receivedData []byte)) {
	for msg := range msgs {
		if c.status == 1 && len(msg.Body) > 0 {
			// 正常客户端状态下处理消息
			callbackFunDealMsg(msg.Body)
		} else if c.status == 0 {
			// 客户端异常状态，关闭连接
			return
		}
	}
}

// OnConnectionError 消费者端，掉线重连失败后的错误回调
func (c *Consumer) OnConnectionError(callbackOfflineErr func(err *amqp.Error)) {
	c.callbackOffLine = callbackOfflineErr
	go c.monitorConnection()
}

func (c *Consumer) monitorConnection() {
	err := <-c.connErr
	c.handleConnectionError(err)
}

func (c *Consumer) handleConnectionError(err *amqp.Error) {
	attempts := 1
	for attempts <= c.config.RetryTimes {
		attempts++
		time.Sleep(c.config.ReconnectInterval * time.Second)
		log.Slog.ErrorF(c.ctx, "RabbitMQ consumer connection error: %s, retry attempt: %d", err, attempts)

		if c.status == 1 {
			c.receivedMsgBlocking <- struct{}{}
		}

		newConsumer, err := CreateConsumer(c.config)
		if err != nil {
			log.Slog.ErrorF(c.ctx, "RabbitMQ consumer connection error: %s", err)
			continue
		}

		c.swapConnection(newConsumer)
		return
	}

	// 如果超过最大重连次数，调用回调函数
	c.callbackOffLine(err)
}

func (c *Consumer) swapConnection(newConsumer *Consumer) {
	c.ctx = newConsumer.ctx
	c.config = newConsumer.config
	c.connect = newConsumer.connect
	c.connErr = newConsumer.connErr
	c.callbackForReceived = newConsumer.callbackForReceived
	c.callbackOffLine = newConsumer.callbackOffLine
	c.enableDelayMsgPlugin = newConsumer.enableDelayMsgPlugin
	c.receivedMsgBlocking = newConsumer.receivedMsgBlocking

	go c.connect.NotifyClose(c.connErr)
	go c.OnConnectionError(c.callbackOffLine)
	c.Received(c.callbackForReceived)
}

// close 关闭连接
func (c *Consumer) close() {
	_ = c.connect.Close()
}

package work_queue

import (
	"context"
	"fmt"
	"github.com/qiafan666/gotato/commons/glog"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

func CreateConsumer(config ConsumerConfig) (*Consumer, error) {

	if config.Ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	if config.ChanNumber <= 0 {
		return nil, fmt.Errorf("channel number is less than 1")
	}

	// 获取配置信息
	conn, err := amqp.Dial(config.Addr)
	if err != nil {
		glog.Slog.ErrorF(nil, "rabbitmq consumer connect error: %s", err.Error())
		return nil, err
	}

	cons := &Consumer{
		ctx:                 config.Ctx,
		config:              config,
		connect:             conn,
		connErr:             conn.NotifyClose(make(chan *amqp.Error, 1)),
		receivedMsgBlocking: make(chan struct{}),
		status:              1,
	}
	return cons, nil
}

// Consumer 定义一个消息队列结构体：WorkQueue 模型
type Consumer struct {
	ctx                 context.Context
	config              ConsumerConfig
	connect             *amqp.Connection
	connErr             chan *amqp.Error
	callbackForReceived func(receivedData []byte) //   断线重连，结构体内部使用
	callbackOffLine     func(err *amqp.Error)     //   断线重连，结构体内部使用
	receivedMsgBlocking chan struct{}             // 接受消息时用于阻塞消息处理函数
	status              byte                      // 客户端状态：1=正常；0=异常
	ch                  *amqp.Channel             // 通道
}

// Received 接收、处理消息
func (c *Consumer) Received(callbackFunDealMsg func(receivedData []byte)) {
	defer c.close() // 关闭连接

	// 存储回调函数，用于可能的掉线重连
	c.callbackForReceived = callbackFunDealMsg

	// 为每个通道创建一个单独的消费者协程
	for i := 1; i <= c.config.ChanNumber; i++ {
		go c.consumeMessagesOnChannel(i, callbackFunDealMsg)
	}

	// 接收到消息或关闭通道
	if _, isOk := <-c.receivedMsgBlocking; isOk {
		c.status = 0
		close(c.receivedMsgBlocking)
	}
}

// consumeMessagesOnChannel 处理单个通道上的消息接收
func (c *Consumer) consumeMessagesOnChannel(chanNo int, callbackFunDealMsg func(receivedData []byte)) {
	ch, err := c.connect.Channel()
	if err != nil {
		glog.Slog.ErrorF(c.ctx, "创建RabbitMQ通道失败: %s, chanNo: %d", err.Error(), chanNo)
		return
	}
	defer ch.Close()

	q, err := c.declareQueue(ch, chanNo)
	if err != nil {
		return // 相关错误已在 declareQueue 中记录
	}

	c.ch = ch

	msgs, err := c.startConsuming(ch, q, chanNo)
	if err != nil {
		return // 错误日志在 startConsuming 中已处理
	}

	c.processReceivedMessages(msgs, callbackFunDealMsg)
}

// declareQueue 声明队列
func (c *Consumer) declareQueue(ch *amqp.Channel, chanNo int) (*amqp.Queue, error) {
	q, err := ch.QueueDeclare(
		c.config.QueueName,
		c.config.Durable,
		true,  // 是否自动删除
		false, // 是否独占
		false, // 是否等待
		nil,   // 其他属性
	)
	if err != nil {
		glog.Slog.ErrorF(c.ctx, "声明队列失败: %s, chanNo: %d", err.Error(), chanNo)
		return nil, err
	}
	return &q, nil
}

// SetQos 设置质量保证
func (c *Consumer) SetQos(prefetchCount int, prefetchSize int, global bool) error {
	err := c.ch.Qos(
		prefetchCount, // 预取计数
		prefetchSize,  // 预取大小
		global,        // 全局应用
	)
	if err != nil {
		glog.Slog.ErrorF(c.ctx, "设置Qos失败: %s", err.Error())
	}
	return err
}

// startConsuming 开始消费消息
func (c *Consumer) startConsuming(ch *amqp.Channel, q *amqp.Queue, chanNo int) (<-chan amqp.Delivery, error) {
	msgs, err := ch.ConsumeWithContext(
		c.ctx,
		q.Name,
		"",    // 消费者标签
		true,  // 自动确认
		false, // 是否独占
		false, // noLocal（不支持）
		false, // 是否等待
		nil,
	)
	if err != nil {
		glog.Slog.ErrorF(c.ctx, "开始消息消费失败: %s, chanNo: %d", err.Error(), chanNo)
		return nil, err
	}
	return msgs, nil
}

// processReceivedMessages 处理接收到的消息
func (c *Consumer) processReceivedMessages(msgs <-chan amqp.Delivery, callbackFunDealMsg func(receivedData []byte)) {
	for msg := range msgs {
		if c.status == 1 && len(msg.Body) > 0 {
			callbackFunDealMsg(msg.Body)
		} else if c.status == 0 {
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
		glog.Slog.ErrorF(c.ctx, "RabbitMQ consumer connection error: %s, retry attempt: %d", err, attempts)

		if c.status == 1 {
			c.receivedMsgBlocking <- struct{}{}
		}

		newConsumer, err := CreateConsumer(c.config)
		if err != nil {
			glog.Slog.ErrorF(c.ctx, "RabbitMQ consumer connection error: %s", err)
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
	c.receivedMsgBlocking = newConsumer.receivedMsgBlocking

	go c.connect.NotifyClose(c.connErr)
	go c.OnConnectionError(c.callbackOffLine)
	c.Received(c.callbackForReceived)
}

// close 关闭连接
func (c *Consumer) close() {
	_ = c.connect.Close()
}

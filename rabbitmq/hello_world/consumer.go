package hello_world

import (
	"context"
	"fmt"
	"github.com/qiafan666/gotato/commons/log"
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
		log.Slog.ErrorF(config.Ctx, "rabbitmq consumer connect error:%v", err)
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

// 定义一个消息队列结构体：helloworld 模型
type Consumer struct {
	ctx                 context.Context
	config              ConsumerConfig
	connect             *amqp.Connection
	connErr             chan *amqp.Error
	callbackForReceived func(receivedData []byte) //   断线重连，结构体内部使用
	callbackOffLine     func(err *amqp.Error)     //   断线重连，结构体内部使用
	receivedMsgBlocking chan struct{}             // 接受消息时用于阻塞消息处理函数
	status              byte                      // 客户端状态：1=正常；0=异常
}

// Received 接收、处理消息
func (c *Consumer) Received(callbackFunDealMsg func(receivedData []byte)) {
	defer c.close() // 确保在函数退出时关闭连接

	// 为重连场景，将回调函数存储到结构体变量中
	c.callbackForReceived = callbackFunDealMsg

	// 启动多个goroutine并发处理消息
	for i := 1; i <= c.config.ChanNumber; i++ {
		go c.handleChannel(i)
	}

	// 阻塞直到接收到停止消费者的信号
	if _, isOk := <-c.receivedMsgBlocking; isOk {
		c.status = 0
		close(c.receivedMsgBlocking)
	}
}

func (c *Consumer) handleChannel(chanNo int) {
	ch, err := c.connect.Channel()
	if err != nil {
		log.Slog.ErrorF(c.ctx, "创建rabbitmq通道失败:%v, 通道编号:%d", err, chanNo)
		return
	}
	defer ch.Close() // 确保在函数退出时关闭通道

	err = c.setQos(ch, chanNo)
	if err != nil {
		return
	}

	// 声明并初始化队列
	queue, err := c.declareQueue(ch)
	if err != nil {
		log.Slog.ErrorF(c.ctx, "声明队列失败:%v", err)
		return
	}

	// 开始消费消息
	msgs, err := c.consumeQueue(ch, queue.Name)
	if err != nil {
		log.Slog.ErrorF(c.ctx, "开始消费消息失败:%v", err)
		return
	}

	// 处理收到的消息
	c.processMessages(msgs, c.callbackForReceived)
}

// setQos 设置质量保证
func (c *Consumer) setQos(ch *amqp.Channel, chanNo int) error {
	err := ch.Qos(
		1,     // 预取计数
		0,     // 预取大小
		false, // 全局应用
	)
	if err != nil {
		log.Slog.ErrorF(c.ctx, "设置Qos失败: %s, chanNo: %d", err.Error(), chanNo)
	}
	return err
}

func (c *Consumer) declareQueue(ch *amqp.Channel) (amqp.Queue, error) {
	return ch.QueueDeclare(
		c.config.QueueName,
		c.config.Durable,
		true,  // 当无人使用时自动删除
		false, // 是否独占
		false, // 是否不等待
		nil,   // 参数
	)
}

func (c *Consumer) consumeQueue(ch *amqp.Channel, queueName string) (<-chan amqp.Delivery, error) {
	return ch.ConsumeWithContext(
		c.ctx,
		queueName,
		"",    // 消费者标签，确保在消息通道中唯一
		true,  // 自动确认
		false, // 是否独占
		false, // no-local标志，不适用于RabbitMQ
		false, // 是否不等待
		nil,   // 参数
	)
}

func (c *Consumer) processMessages(msgs <-chan amqp.Delivery, callbackFunDealMsg func(receivedData []byte)) {
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
	c.receivedMsgBlocking = newConsumer.receivedMsgBlocking

	go c.connect.NotifyClose(c.connErr)
	go c.OnConnectionError(c.callbackOffLine)
	c.Received(c.callbackForReceived)
}

// close 关闭连接
func (c *Consumer) close() {
	_ = c.connect.Close()
}

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

// 定义一个消息队列结构体：PublishSubscribe 模型
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
}

// Received 接收、处理消息
func (c *Consumer) Received(callbackFunDealMsg func(receivedData []byte)) {
	defer func() {
		c.close()
	}()

	// 将回调函数地址赋值给结构体变量，用于掉线重连使用
	c.callbackForReceived = callbackFunDealMsg

	for i := 1; i <= c.config.ChanNumber; i++ {
		go func(chanNo int) {

			ch, err := c.connect.Channel()
			if err != nil {
				log.Slog.ErrorF(c.ctx, "rabbitmq consumer connect Channel error:%v, chanNo:%d", err, chanNo)
				return
			}
			defer func() {
				_ = ch.Close()
			}()

			// 声明exchange交换机
			err = ch.ExchangeDeclare(
				c.config.ExchangeName, //exchange name
				c.config.ExchangeType, //exchange kind
				c.config.Durable,      //数据是否持久化
				!c.config.Durable,     //所有连接断开时，交换机是否删除
				false,
				false,
				nil,
			)
			if err != nil {
				log.Slog.ErrorF(c.ctx, "rabbitmq consumer connect ExchangeDeclare error:%v", err)
				return
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
				log.Slog.ErrorF(c.ctx, "rabbitmq consumer connect QueueDeclare error:%v", err)
				return
			}

			//队列绑定
			err = ch.QueueBind(
				queue.Name,
				"", //direct key，  fanout 模式设置为 空 即可
				c.config.ExchangeName,
				false,
				nil,
			)
			if err != nil {
				log.Slog.ErrorF(c.ctx, "rabbitmq consumer connect QueueBind error:%v", err)
				return
			}

			msgs, err := ch.ConsumeWithContext(
				c.ctx,
				queue.Name, // 队列名称
				"",         //  消费者标记，请确保在一个消息频道唯一
				true,       //是否自动确认，这里设置为 true，自动确认
				false,      //是否私有队列，false标识允许多个 consumer 向该队列投递消息，true 表示独占
				false,      //RabbitMQ不支持noLocal标志。
				false,      // 队列如果已经在服务器声明，设置为 true ，否则设置为 false；
				nil,
			)
			if err != nil {
				log.Slog.ErrorF(c.ctx, "rabbitmq consumer connect Consume error:%v", err)
				return
			}

			for {
				select {
				case msg := <-msgs:
					// 消息处理
					if c.status == 1 && len(msg.Body) > 0 {
						callbackFunDealMsg(msg.Body)
					} else if c.status == 0 {
						return
					}
				}
			}
		}(i)
	}
	if _, isOk := <-c.receivedMsgBlocking; isOk {
		c.status = 0
		close(c.receivedMsgBlocking)
	}

}

// OnConnectionError 消费者端，掉线重连失败后的错误回调
func (c *Consumer) OnConnectionError(callbackOfflineErr func(err *amqp.Error)) {
	c.callbackOffLine = callbackOfflineErr
	go func() {
		select {
		case err := <-c.connErr:
			var i = 1
			for i = 1; i <= c.config.RetryTimes; i++ {
				// 自动重连机制
				time.Sleep(c.config.ReconnectInterval * time.Second)
				log.Slog.ErrorF(c.ctx, "rabbitmq consumer connect error:%v, retry times:%d", err, i)
				// 发生连接错误时,中断原来的消息监听（包括关闭连接）
				if c.status == 1 {
					c.receivedMsgBlocking <- struct{}{}
				}
				conn, err := CreateConsumer(c.config)
				if err != nil {
					continue
				} else {
					go func() {
						c.connErr = conn.connect.NotifyClose(make(chan *amqp.Error, 1))
						go conn.OnConnectionError(c.callbackOffLine)
						conn.Received(c.callbackForReceived)
					}()
					// 新的客户端重连成功后，释放旧的回调函数 - OnConnectionError
					if c.status == 0 {
						return
					}
					break
				}
			}
			if i > c.config.RetryTimes {
				callbackOfflineErr(err)
				// 如果超过最大重连次数，同样需要释放回调函数 - OnConnectionError
				if c.status == 0 {
					return
				}
			}
		}
	}()
}

// close 关闭连接
func (c *Consumer) close() {
	_ = c.connect.Close()
}

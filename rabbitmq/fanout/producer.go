package fanout

import (
	"context"
	"github.com/qiafan666/gotato/commons/log"
	amqp "github.com/rabbitmq/amqp091-go"
)

// CreateProducer 创建一个生产者
func CreateProducer(options ...OptionsProd) (*producer, error) {
	// 获取配置信息
	conn, err := amqp.Dial(RabbitMqAddr)

	if err != nil {
		log.Slog.ErrorF(nil, "rabbitmq producer connect error: %s", err.Error())
		return nil, err
	}

	prod := &producer{
		connect:      conn,
		exchangeType: RabbitMqExchangeType,
		exchangeName: RabbitMqExchangeName,
		queueName:    RabbitMqQueueName,
		durable:      RabbitMqDurable,
		args:         nil,
	}
	// 加载用户设置的参数
	for _, val := range options {
		val.apply(prod)
	}
	return prod, nil
}

// 定义一个消息队列结构体：PublishSubscribe 模型
type producer struct {
	connect              *amqp.Connection
	exchangeType         string
	exchangeName         string
	queueName            string
	durable              bool
	occurError           error
	enableDelayMsgPlugin bool
	args                 amqp.Table
}

// Send 发送消息
// 参数：
// data 发送的数据、
// delayMillisecond 延迟时间(毫秒)，只有启用了消息延迟插件才有效果
func (p *producer) Send(data string, delayMillisecond int) bool {

	// 获取一个频道
	ch, err := p.connect.Channel()
	p.occurError = err
	defer func() {
		_ = ch.Close()
	}()

	// 声明交换机，该模式生产者只负责将消息投递到交换机即可
	err = ch.ExchangeDeclare(
		p.exchangeName, //交换器名称
		p.exchangeType, //fanout 模式(扇形模式，发布/订阅 模式) ，解决 发布、订阅场景相关的问题
		p.durable,      //durable
		!p.durable,     //autodelete
		false,
		false,
		p.args,
	)
	p.occurError = err

	// 如果队列的声明是持久化的，那么消息也设置为持久化
	msgPersistent := amqp.Transient
	if p.durable {
		msgPersistent = amqp.Persistent
	}
	// 投递消息
	if err == nil {
		err = ch.Publish(
			p.exchangeName, // 交换机名称
			p.queueName,    // fanout 模式默认为空，表示所有订阅的消费者会接受到相同的消息
			false,
			false,
			amqp.Publishing{
				DeliveryMode: msgPersistent, //消息是否持久化，这里与保持保持一致即可
				ContentType:  "text/plain",
				Body:         []byte(data),
				Headers: amqp.Table{
					"x-delay": delayMillisecond, // 延迟时间: 毫秒
				},
			})
	}

	p.occurError = err
	if p.occurError != nil { //  发生错误，返回 false
		log.Slog.ErrorF(context.Background(), "rabbitmq producer send error: %s", p.occurError.Error())
		return false
	} else {
		return true
	}
}

// Close 发送完毕手动关闭，这样不影响send多次发送数据
func (p *producer) Close() {
	_ = p.connect.Close()
}

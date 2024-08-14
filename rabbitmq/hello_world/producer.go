package hello_world

import (
	"context"
	"fmt"
	"github.com/qiafan666/gotato/commons/glog"
	amqp "github.com/rabbitmq/amqp091-go"
)

// CreateProducer 创建一个生产者
func CreateProducer(config ProducerConfig) (*Producer, error) {

	if config.Ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	// 获取配置信息
	conn, err := amqp.Dial(config.Addr)

	if err != nil {
		glog.Slog.ErrorF(nil, "rabbitmq producer connect error: %s", err.Error())
		return nil, err
	}

	prod := &Producer{
		ctx:     config.Ctx,
		config:  config,
		connect: conn,
	}
	return prod, nil
}

// Producer 定义一个消息队列结构体：helloworld 模型
type Producer struct {
	ctx     context.Context
	config  ProducerConfig
	connect *amqp.Connection
}

func (p *Producer) Send(data []byte) bool {

	// 获取一个通道
	ch, err := p.connect.Channel()
	if err != nil {
		glog.Slog.ErrorF(p.ctx, "rabbitmq channel error: %s", err.Error())
		return false
	}

	defer func() {
		_ = ch.Close()
	}()

	// 声明消息队列
	_, err = ch.QueueDeclare(
		p.config.QueueName, // 队列名称
		p.config.Durable,   //是否持久化，false模式数据全部处于内存，true会保存在erlang自带数据库，但是影响速度
		!p.config.Durable,  //生产者、消费者全部断开时是否删除队列。一般来说，数据需要持久化，就不删除；非持久化，就删除
		false,              //是否私有队列，false标识允许多个 consumer 向该队列投递消息，true 表示独占
		false,              // 队列如果已经在服务器声明，设置为 true ，否则设置为 false；
		nil,                // 相关参数
	)
	if err != nil {
		glog.Slog.ErrorF(p.ctx, "rabbitmq queue declare error: %s", err.Error())
		return false
	}

	// 如果队列的声明是持久化的，那么消息也设置为持久化
	msgPersistent := amqp.Transient
	if p.config.Durable {
		msgPersistent = amqp.Persistent
	}
	// 投递消息
	err = ch.PublishWithContext(
		p.ctx,
		"",                 // helloworld 、workqueue 模式设置为空字符串，表示使用默认交换机
		p.config.QueueName, //  direct key，注意：简单模式与队列名称相同
		false,
		false,
		amqp.Publishing{
			DeliveryMode: msgPersistent, //消息是否持久化，这里与保持保持一致即可
			ContentType:  "text/plain",
			Body:         data,
		})
	if err != nil {
		glog.Slog.ErrorF(p.ctx, "rabbitmq publish error: %s", err.Error())
		return false
	}
	return true
}

// Close 发送完毕手动关闭，这样不影响send多次发送数据
func (p *Producer) Close() {
	_ = p.connect.Close()
}

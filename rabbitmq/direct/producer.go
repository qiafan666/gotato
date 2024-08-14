package direct

import (
	"context"
	"fmt"
	"github.com/qiafan666/gotato/commons/glog"
	amqp "github.com/rabbitmq/amqp091-go"
)

// CreateProducer  创建一个生产者
func CreateProducer(config ProducerConfig, options ...OptionsProd) (*Producer, error) {

	if config.Ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	// 获取配置信息
	conn, err := amqp.Dial(config.Addr)
	if err != nil {
		glog.Slog.ErrorF(config.Ctx, "rabbitmq producer connect error: %s", err.Error())
		return nil, err
	}

	prod := &Producer{
		ctx:     config.Ctx,
		config:  config,
		connect: conn,
		args:    config.Args,
	}
	// 加载用户设置的参数
	for _, val := range options {
		val.apply(prod)
	}
	return prod, nil
}

// Producer 定义一个消息队列结构体：Routing 模型
type Producer struct {
	ctx                  context.Context
	config               ProducerConfig
	connect              *amqp.Connection
	enableDelayMsgPlugin bool // 是否使用延迟队列模式
	args                 amqp.Table
}

// Send 发送消息
// 参数：
// routeKey 路由键、
// data 发送的数据、
// delayMillisecond 延迟时间(毫秒)，只有启用了消息延迟插件才有效果
func (p *Producer) Send(routeKey string, data []byte, delayMillisecond int) bool {

	// 获取一个频道
	ch, err := p.connect.Channel()
	if err != nil {
		glog.Slog.ErrorF(p.ctx, "rabbitmq channel error: %s", err.Error())
		return false
	}
	defer func() {
		_ = ch.Close()
	}()

	// 声明交换机，该模式生产者只负责将消息投递到交换机即可
	err = ch.ExchangeDeclare(
		p.config.ExchangeName, //交换器名称
		p.config.ExchangeType, //direct(定向消息), 按照路由键名匹配消息
		p.config.Durable,      //消息是否持久化
		!p.config.Durable,     //交换器是否自动删除
		false,
		false,
		p.args,
	)
	if err != nil {
		glog.Slog.ErrorF(p.ctx, "rabbitmq exchange declare error: %s", err.Error())
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
		p.config.ExchangeName, // 交换机名称
		routeKey,              // direct 模式默认为空即可
		false,
		false,
		amqp.Publishing{
			DeliveryMode: msgPersistent, //消息是否持久化，这里与保持保持一致即可
			ContentType:  "text/plain",
			Body:         data,
			Headers: amqp.Table{
				"x-delay": delayMillisecond, // 延迟时间: 毫秒
			},
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

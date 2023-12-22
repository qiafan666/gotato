package hello_world

import (
	"github.com/qiafan666/gotato/commons/log"
	amqp "github.com/rabbitmq/amqp091-go"
)

// CreateProducer 创建一个生产者
func CreateProducer(config ProducerConfig) (*Producer, error) {
	// 获取配置信息
	conn, err := amqp.Dial(config.Addr)

	if err != nil {
		log.Slog.ErrorF(nil, "rabbitmq producer connect error: %s", err.Error())
		return nil, err
	}

	prod := &Producer{
		config:  config,
		connect: conn,
	}
	return prod, nil
}

// 定义一个消息队列结构体：helloworld 模型
type Producer struct {
	config     ProducerConfig
	connect    *amqp.Connection
	occurError error
}

func (p *Producer) Send(data string) bool {

	// 获取一个通道
	ch, err := p.connect.Channel()
	p.occurError = err

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
	p.occurError = err

	// 如果队列的声明是持久化的，那么消息也设置为持久化
	msgPersistent := amqp.Transient
	if p.config.Durable {
		msgPersistent = amqp.Persistent
	}
	// 投递消息
	if err == nil {
		err = ch.Publish(
			"",                 // helloworld 、workqueue 模式设置为空字符串，表示使用默认交换机
			p.config.QueueName, //  direct key，注意：简单模式与队列名称相同
			false,
			false,
			amqp.Publishing{
				DeliveryMode: msgPersistent, //消息是否持久化，这里与保持保持一致即可
				ContentType:  "text/plain",
				Body:         []byte(data),
			})
	}
	p.occurError = err
	if p.occurError != nil { //  发生错误，返回 false
		log.Slog.ErrorF(nil, "rabbitmq Send error: %s", p.occurError.Error())
		return false
	} else {
		return true
	}
}

// Close 发送完毕手动关闭，这样不影响send多次发送数据
func (p *Producer) Close() {
	_ = p.connect.Close()
}

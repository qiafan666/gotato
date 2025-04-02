package gkafka

type IHandler interface {
	Handle(topic, msg string)
}

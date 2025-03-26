package kafka

type IHandler interface {
	Handle(topic, msg string)
}

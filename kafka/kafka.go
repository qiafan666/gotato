package kafka

type MsgHandler interface {
	Handle(topic, msg string)
}

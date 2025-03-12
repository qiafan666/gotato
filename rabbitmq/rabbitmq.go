package rabbitmq

type MsgHandler interface {
	Handle(msg string) error
}

package rabbitmq

type IHandler interface {
	Handle(msg string) error
}

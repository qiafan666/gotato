package grabbitmq

type IHandler interface {
	Handle(msg string) error
}

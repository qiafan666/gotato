package grocketmq

type IHandler interface {
	Handle(tags, msg string) error
}

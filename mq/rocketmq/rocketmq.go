package rocketmq

type IHandler interface {
	Handle(tags, msg string) error
}

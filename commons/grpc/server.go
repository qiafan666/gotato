package grpc

type IHandler interface {
	Handle(request *Message, ch chan<- *Message)
}

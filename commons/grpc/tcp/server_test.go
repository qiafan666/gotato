package tcp

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/qiafan666/gotato/commons/gface"
	"github.com/qiafan666/gotato/commons/gid"
	"github.com/qiafan666/gotato/commons/grpc"
	"github.com/qiafan666/gotato/commons/gson"
	"testing"
	"time"
)

type Handler interface {
	Handle(*grpc.Message) *grpc.Message
}

type Router struct {
	routes map[grpc.Command]Handler
}

func NewRouter() *Router {
	r := &Router{routes: make(map[grpc.Command]Handler)}
	r.Register(grpc.CmdTestLogic, &Test{})
	return r
}

func (r *Router) Register(cmd grpc.Command, handler Handler) {
	r.routes[cmd] = handler
}

func (r *Router) Handle(msg *grpc.Message, out chan<- *grpc.Message) {
	if handler, ok := r.routes[msg.Command]; ok {
		resp := handler.Handle(msg)
		out <- resp
	}
}

func TestServer(t *testing.T) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 20*time.Minute)

	server := NewServer(":10081", NewRouter(), &ServerOptions{
		Timeout: 3 * time.Second,
		Logger:  gface.NewLogger("server", zapLog()),
	})
	server.Run(ctx)

	<-ctx.Done()
	cancelFunc()
}

func TestClient1(t *testing.T) {

	addr := "127.0.0.1:10081"
	cmd := grpc.CmdTestLogic
	data := []any{"BTCUSDT", 2, 0}

	client := NewClient(context.Background(), addr, &ClientOptions{
		MaxConn:    2,
		IdleConn:   0,
		Timeout:    10 * time.Second,
		RetryLimit: 2,
		Logger:     gface.NewLogger("client", zapLog()),
		Hystrix: HystrixOptions{
			Timeout:                5000 * time.Millisecond,
			SleepWindow:            500 * time.Millisecond,
			MaxConcurrentRequests:  5000,
			RequestVolumeThreshold: 100,
			ErrorPercentThreshold:  50,
		},
	})
	marshal, _ := gson.Marshal(data)
	for i := 0; i < 10; i++ {
		time.Sleep(2 * time.Second)
		client.Do(context.Background(), &grpc.Message{
			Command: cmd,
			PkgType: grpc.PkgTypeRequest,
			ReqId:   gid.RandID(),
			Seq:     grpc.NewSeq(),
			Body:    marshal,
		})
	}

}

type Test struct{}

func (t *Test) Handle(request *grpc.Message) (resp *grpc.Message) {
	params := make([]any, 0)
	var market string
	var side uint32
	var off uint32
	params = append(params, market, side, off)
	gson.Unmarshal(request.Body, &params)
	fmt.Println(params)
	result := [2]any{
		map[string]any{
			"reqId": request.ReqId,
			"time":  time.Now().UTC().Format(time.RFC3339Nano),
		},
		[]any{
			request.ReqId,
			time.Now().UTC().Format(time.RFC3339Nano),
		},
	}
	marshal, _ := json.Marshal(result[request.Seq%2])

	response := &grpc.Message{
		Command: request.Command,
		PkgType: grpc.PkgTypeReply,
		ReqId:   request.ReqId,
		Seq:     request.Seq,
		Result:  0,
		Body:    marshal,
	}
	return response
}

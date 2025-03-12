package tcp

import (
	"context"
	"encoding/json"
	"github.com/qiafan666/gotato/commons/grpc"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 20*time.Minute)

	server := NewServer(":10081", new(testServerHandler), &ServerOptions{
		Timeout: 3 * time.Second,
	})
	server.Run(ctx)

	//client
	//addr := "192.168.1.156:27416"
	addr := "127.0.0.1:10081"
	cmd := grpc.CmdTestLogic
	data := []any{"BTCUSDT", 2, 0}

	//addr := "192.168.31.103:7316"
	//cmd := 701
	//data := []any{}

	client := NewClient(context.Background(), addr, &ClientOptions{
		MaxConn:    2,
		IdleConn:   0,
		Timeout:    10 * time.Second,
		RetryLimit: 2,
	})
	marshal, _ := json.Marshal(data)

	client.Do(context.Background(), &grpc.Message{
		Command:  cmd,
		PkgType:  grpc.PkgTypeRequest,
		ReqId:    uint64(1),
		Sequence: uint32(1),
		Body:     marshal,
	})

	<-ctx.Done()
	cancelFunc()
}

type testServerHandler struct{}

func (h *testServerHandler) Handle(request *grpc.Message, ch chan<- *grpc.Message) {
	params := make([]any, 0)
	var market string
	var side uint32
	var off uint32
	params = append(params, &market, &side, &off)
	json.Unmarshal(request.Body, &params)
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
	marshal, _ := json.Marshal(result[request.Sequence%2])

	response := &grpc.Message{
		Command:  request.Command,
		PkgType:  grpc.PkgTypeReply,
		ReqId:    request.ReqId,
		Sequence: request.Sequence,
		Result:   0,
		Body:     marshal,
	}

	ch <- response
}

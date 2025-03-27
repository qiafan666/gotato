package tcp

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"github.com/qiafan666/gotato/commons/gconc"
	"github.com/qiafan666/gotato/commons/gface"
	"github.com/qiafan666/gotato/commons/gid"
	"github.com/qiafan666/gotato/commons/grpc"
	"github.com/qiafan666/gotato/commons/grpc/tcp/protocol"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	//addr := "192.168.1.156:27416"
	addr := "10.254.81.138:8080"
	//addr := "10.254.223.207:7316"
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
	_ = client

	Seq := gid.NewSerialId[uint32]()
	wg := gconc.NewWaitGroup()
	for i := 0; i < 10; i++ {
		i := i
		wg.Go(func() {
			marshal, _ := json.Marshal(data)

			resp, err := client.Do(context.Background(), &grpc.Message{
				Command: cmd,
				PkgType: grpc.PkgTypeRequest,
				ReqId:   int64(i + 1),
				Seq:     Seq.Id(),
				Body:    marshal,
			})

			assert.Equal(t, nil, err)
			//t.Logf("%+v", resp)
			t.Logf("%s", resp.Body)
			<-time.After(2 * time.Second)
		})
	}

	wg.Wait()
}

func TestWrite(t *testing.T) {
	conn, err := net.Dial("tcp", "192.168.1.156:27416")
	assert.Nil(t, err, err)

	str := "62656570950100000000000000008aabee2a02000000e626e5ef81572c6d0f00000000005b2242544355534454222c322c305d"
	data, _ := hex.DecodeString(str)

	conn.Write(data)

	ctx := context.Background()
	p := protocol.New()
	resp, e := p.Decode(ctx, conn)
	assert.Equal(t, nil, e)

	conn.Close()

	t.Logf("%s", resp.Body)
}

func TestHeartbeat(t *testing.T) {
	//addr := "192.168.1.156:27416"
	addr := "127.0.0.1:10081"
	cmd := grpc.CmdHeartbeat
	timeout := uint32(5000)

	client := NewClient(context.Background(), addr, &ClientOptions{
		MaxConn:    1,
		IdleConn:   0,
		Timeout:    10 * time.Second,
		RetryLimit: 2,
		Logger:     gface.NewLogger("testHeartbeat", zapLog()),
	})
	_ = client

	wg := gconc.NewWaitGroup()
	for i := 0; i < 1; i++ {
		wg.Go(func() {
			resp, err := client.Do(context.Background(), &grpc.Message{
				Command: cmd,
				PkgType: grpc.PkgTypeRequest,
				Heartbeat: &grpc.Heartbeat{
					Timeout: timeout,
				},
			})
			assert.Equal(t, nil, err)
			assert.Nil(t, resp) // 心跳包不需要等待响应 resp==nil
			//t.Logf("%s", resp.Body)
			<-time.After(2 * time.Second)
		})
	}
	wg.Wait()
}

package tcp

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/qiafan666/gotato/commons/grpc"
	"github.com/stretchr/testify/assert"
	"net"
	"sync"
	"testing"
	"time"
)

func TestConn(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)

	go func() {
		for j := 0; j < 5; j++ {
			conn, err := net.DialTimeout("tcp", "192.168.1.156:27416", 5*time.Second)
			assert.Nil(t, err, err)
			c := NewConn(ctx, conn, nil)

			go func() {
				for i := 0; i < 10; i++ {
					cmd := grpc.CmdTestLogic
					data := []any{"BTCUSDT", 2, 0}
					marshal, _ := json.Marshal(data)
					Seq := grpc.NewSeq()
					v := &grpc.Message{
						Command:   cmd,
						PkgType:   grpc.PkgTypeRequest,
						ReqId:     time.Now().UnixNano(),
						Seq:       Seq,
						Result:    0,
						Body:      marshal,
						Heartbeat: nil,
					}
					ch := NewRecvChan(fmt.Sprintf("%d", Seq))
					e := c.Send(v, ch)
					assert.Equal(t, nil, e)

					resp := <-ch.Ch
					ch.Close()
					t.Logf("conn:%d, Seq:%d, body:%s", c.connId, resp.Seq, resp.Body)
					<-time.After(500 * time.Millisecond)
				}
			}()
		}
	}()

	select {
	case <-ctx.Done():
	}
	<-time.After(1 * time.Second)
}

func TestNetPollConn(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	conn, err := net.DialTimeout("tcp", "192.168.1.156:27416", 5*time.Second)
	//addr, _ := netpoll.ResolveTCPAddr("tcp", "192.168.1.156:27416")
	//conn, err := netpoll.DialTCP(ctx, "tcp", nil, addr)
	assert.Nil(t, err, err)
	c := NewConn(ctx, conn, nil)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			cmd := grpc.CmdTestLogic
			data := []any{"BTCUSDT", 2, 0}
			marshal, _ := json.Marshal(data)
			Seq := grpc.NewSeq()
			v := &grpc.Message{
				Command:   cmd,
				PkgType:   grpc.PkgTypeRequest,
				ReqId:     time.Now().UnixNano(),
				Seq:       Seq,
				Result:    0,
				Body:      marshal,
				Heartbeat: nil,
			}
			ch := NewRecvChan(fmt.Sprintf("%d", Seq))
			e := c.Send(v, ch)
			assert.Equal(t, nil, e)

			resp := <-ch.Ch
			ch.Close()
			if resp == nil {
				t.Logf("channel closed")
				continue
			}
			t.Logf("conn:%d, Seq:%d, body:%s", c.connId, resp.Seq, resp.Body)
			<-time.After(500 * time.Millisecond)
		}
	}()

	wg.Wait()
	time.AfterFunc(3*time.Second, cancel)
	<-time.After(100 * time.Millisecond)
}

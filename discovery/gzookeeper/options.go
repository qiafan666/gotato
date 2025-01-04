package gzookeeper

import (
	"github.com/qiafan666/gotato/commons/gcommon"
	"google.golang.org/grpc"
	"time"
)

var GetZkCtx = gcommon.SetTraceId("gzookeeper")

type ZkOption func(*ZkClient)

func WithRoundRobin() ZkOption {
	return func(client *ZkClient) {
		client.balancerName = "round_robin"
	}
}

func WithUserNameAndPassword(userName, password string) ZkOption {
	return func(client *ZkClient) {
		client.username = userName
		client.password = password
	}
}

func WithOptions(opts ...grpc.DialOption) ZkOption {
	return func(client *ZkClient) {
		client.options = opts
	}
}

func WithFreq(freq time.Duration) ZkOption {
	return func(client *ZkClient) {
		client.ticker = time.NewTicker(freq)
	}
}

func WithTimeout(timeout int) ZkOption {
	return func(client *ZkClient) {
		client.timeout = timeout
	}
}

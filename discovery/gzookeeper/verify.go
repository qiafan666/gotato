package gzookeeper

import (
	"context"
	"github.com/go-zookeeper/zk"
	"github.com/qiafan666/gotato/commons/gerr"
	"google.golang.org/grpc"
	"sync"
	"time"
)

func Check(ctx context.Context, ZkServers []string, scheme string, options ...ZkOption) error {
	client := &ZkClient{
		ZkServers:  ZkServers,
		zkRoot:     "/",
		scheme:     scheme,
		timeout:    timeout,
		localConns: make(map[string][]*grpc.ClientConn),
		resolvers:  make(map[string]*Resolver),
		lock:       &sync.Mutex{},
	}
	for _, option := range options {
		option(client)
	}

	// 建立与Zookeeper服务器的连接，并设置超时时间，并处理认证。
	conn, eventChan, err := zk.Connect(ZkServers, time.Duration(client.timeout)*time.Second)
	if err != nil {
		return gerr.WrapMsg(err, "connect failed", "ZkServers", ZkServers)
	}

	_, cancel := context.WithCancel(context.Background())
	client.cancel = cancel
	client.ticker = time.NewTicker(defaultFreq)

	// 如果有用户名和密码，则进行认证。
	if client.username != "" && client.password != "" {
		auth := []byte(client.username + ":" + client.password)
		if err = conn.AddAuth("digest", auth); err != nil {
			conn.Close()
			return gerr.WrapMsg(err, "AddAuth failed", "userName", client.username, "password", client.password)
		}
	}

	client.zkRoot += scheme
	client.eventChan = eventChan
	client.conn = conn

	// 验证根节点是否存在，并创建缺失的节点。
	if err = client.ensureRoot(); err != nil {
		conn.Close()
		return gerr.WrapMsg(err, "ensureRoot failed", "zkRoot", client.zkRoot)
	}
	return nil
}

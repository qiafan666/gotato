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
		logger:     nilLog{},
	}
	for _, option := range options {
		option(client)
	}

	// Establish a Zookeeper connection with a specified timeout and handle authentication.
	conn, eventChan, err := zk.Connect(ZkServers, time.Duration(client.timeout)*time.Second)
	if err != nil {
		return gerr.WrapMsg(err, "connect failed", "ZkServers", ZkServers)
	}

	_, cancel := context.WithCancel(context.Background())
	client.cancel = cancel
	client.ticker = time.NewTicker(defaultFreq)

	// Ensure authentication is set if credentials are provided.
	if client.username != "" && client.password != "" {
		auth := []byte(client.username + ":" + client.password)
		if err := conn.AddAuth("digest", auth); err != nil {
			conn.Close()
			return gerr.WrapMsg(err, "AddAuth failed", "userName", client.username, "password", client.password)
		}
	}

	client.zkRoot += scheme
	client.eventChan = eventChan
	client.conn = conn

	// Verify root node existence and create if missing.
	if err := client.ensureRoot(); err != nil {
		conn.Close()
		return gerr.WrapMsg(err, "ensureRoot failed", "zkRoot", client.zkRoot)
	}
	return nil
}

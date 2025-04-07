package gzookeeper

import (
	"context"
	"errors"
	"github.com/go-zookeeper/zk"
	"github.com/qiafan666/gotato/commons/gerr"
	"github.com/qiafan666/gotato/commons/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"net"
	"strconv"
	"sync"
	"time"
)

const (
	defaultFreq = time.Minute * 30
	timeout     = 5
)

type ZkClient struct {
	ZkServers []string
	zkRoot    string
	username  string
	password  string

	rpcRegisterName string
	rpcRegisterAddr string
	isRegistered    bool
	scheme          string

	timeout   int
	conn      *zk.Conn
	eventChan <-chan zk.Event
	node      string
	ticker    *time.Ticker

	lock    sync.Locker
	options []grpc.DialOption

	resolvers           map[string]*Resolver
	localConns          map[string][]*grpc.ClientConn
	cancel              context.CancelFunc
	isStateDisconnected bool
	balancerName        string
}

// NewZkClient initializes a new ZkClient with provided options and establishes a Zookeeper connection.
func NewZkClient(ZkServers []string, scheme string, options ...ZkOption) (*ZkClient, error) {
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

	// 建立zk连接
	conn, eventChan, err := zk.Connect(ZkServers, time.Duration(client.timeout)*time.Second)
	if err != nil {
		return nil, gerr.WrapMsg(err, "connect failed", "ZkServers", ZkServers)
	}

	ctx, cancel := context.WithCancel(context.Background())
	client.cancel = cancel
	client.ticker = time.NewTicker(defaultFreq)

	if client.username != "" && client.password != "" {
		auth := []byte(client.username + ":" + client.password)
		if err = conn.AddAuth("digest", auth); err != nil {
			conn.Close()
			return nil, gerr.WrapMsg(err, "AddAuth failed", "username", client.username, "password", client.password)
		}
	}

	client.zkRoot += scheme
	client.eventChan = eventChan
	client.conn = conn

	if err = client.ensureRoot(); err != nil {
		conn.Close()
		return nil, err
	}

	resolver.Register(client)
	go client.refresh(ctx)
	go client.watch(ctx)

	return client, nil
}

func (s *ZkClient) Close() {
	glog.Slog.InfoKVs(GetZkCtx, "zk close called")
	s.cancel()
	s.ticker.Stop()
	s.conn.Close()
}

func (s *ZkClient) ensureAndCreate(node string) error {
	exists, _, err := s.conn.Exists(node)
	if err != nil {
		return gerr.WrapMsg(err, "Exists failed", "node", node)
	}
	if !exists {
		_, err = s.conn.Create(node, []byte(""), 0, zk.WorldACL(zk.PermAll))
		if err != nil && !errors.Is(err, zk.ErrNodeExists) {
			return gerr.WrapMsg(err, "Create failed", "node", node)
		}
	}
	return nil
}

func (s *ZkClient) refresh(ctx context.Context) {
	for range s.ticker.C {
		glog.Slog.InfoKVs(GetZkCtx, "zk refresh called")
		s.lock.Lock()
		for rpcName := range s.resolvers {
			s.flushResolver(rpcName)
		}
		for rpcName := range s.localConns {
			delete(s.localConns, rpcName)
		}
		s.lock.Unlock()
		glog.Slog.DebugKVs(GetZkCtx, "zk refresh local conns success")
	}
}

func (s *ZkClient) flushResolverAndDeleteLocal(serviceName string) {
	glog.Slog.DebugKVs(GetZkCtx, "zk flush resolver and delete local", "serviceName", serviceName)
	s.flushResolver(serviceName)
	delete(s.localConns, serviceName)
}

func (s *ZkClient) flushResolver(serviceName string) {
	r, ok := s.resolvers[serviceName]
	if ok {
		r.ResolveNowZK(resolver.ResolveNowOptions{})
	}
}

func (s *ZkClient) GetZkConn() *zk.Conn {
	return s.conn
}

func (s *ZkClient) GetRootPath() string {
	return s.zkRoot
}

func (s *ZkClient) GetNode() string {
	return s.node
}

func (s *ZkClient) ensureRoot() error {
	return s.ensureAndCreate(s.zkRoot)
}

func (s *ZkClient) ensureName(rpcRegisterName string) error {
	return s.ensureAndCreate(s.getPath(rpcRegisterName))
}

func (s *ZkClient) getPath(rpcRegisterName string) string {
	return s.zkRoot + "/" + rpcRegisterName
}

func (s *ZkClient) getAddr(host string, port int) string {
	return net.JoinHostPort(host, strconv.Itoa(port))
}

func (s *ZkClient) AddOption(opts ...grpc.DialOption) {
	s.options = append(s.options, opts...)
}

func (s *ZkClient) GetClientLocalConns() map[string][]*grpc.ClientConn {
	return s.localConns
}

package gzookeeper

import (
	"context"
	"fmt"
	"github.com/go-zookeeper/zk"
	"github.com/qiafan666/gotato/commons/gerr"
	"github.com/qiafan666/gotato/service/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"strings"
)

func (s *ZkClient) watch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			glog.Slog.InfoKVs(GetZkCtx, "zk watch ctx done")
			return
		case event := <-s.eventChan:
			glog.Slog.DebugKVs(GetZkCtx, "zk eventChan recv new event", "event", event)
			switch event.Type {
			case zk.EventSession:
				switch event.State {
				case zk.StateHasSession:
					if s.isRegistered && !s.isStateDisconnected {
						glog.Slog.DebugKVs(GetZkCtx, "zk session event stateHasSession, client already registered", "event", event)
						node, err := s.CreateTempNode(s.rpcRegisterName, s.rpcRegisterAddr)
						if err != nil {
							glog.Slog.ErrorKVs(GetZkCtx, "zk session event stateHasSession, create temp node error", err, "event", event)
						} else {
							s.node = node
						}
					}
				case zk.StateDisconnected:
					s.isStateDisconnected = true
				case zk.StateConnected:
					s.isStateDisconnected = false
				default:
					glog.Slog.DebugKVs(GetZkCtx, "zk session event", "event", event)
				}
			case zk.EventNodeChildrenChanged:
				glog.Slog.DebugKVs(GetZkCtx, "zk eventNodeChildrenChanged", "event", event)
				l := strings.Split(event.Path, "/")
				if len(l) > 1 {
					serviceName := l[len(l)-1]
					s.lock.Lock()
					s.flushResolverAndDeleteLocal(serviceName)
					s.lock.Unlock()
				}
				glog.Slog.DebugKVs(GetZkCtx, "zk event handle success", "path", event.Path)
			case zk.EventNodeDataChanged:
			case zk.EventNodeCreated:
				glog.Slog.DebugKVs(GetZkCtx, "zk node create event", "event", event)
			case zk.EventNodeDeleted:
			case zk.EventNotWatching:
			}
		}
	}
}

func (s *ZkClient) GetConnsRemote(ctx context.Context, serviceName string) (conns []resolver.Address, err error) {
	err = s.ensureName(serviceName)
	if err != nil {
		return nil, err
	}

	path := s.getPath(serviceName)
	_, _, _, err = s.conn.ChildrenW(path)
	if err != nil {
		return nil, gerr.WrapMsg(err, "children watch error", "path", path)
	}
	childNodes, _, err := s.conn.Children(path)
	if err != nil {
		return nil, gerr.WrapMsg(err, "get children error", "path", path)
	} else {
		for _, child := range childNodes {
			fullPath := path + "/" + child
			data, _, err := s.conn.Get(fullPath)
			if err != nil {
				return nil, gerr.WrapMsg(err, "get children error", "fullPath", fullPath)
			}
			glog.Slog.DebugKVs(GetZkCtx, "get addr from remote", "conn", string(data))
			conns = append(conns, resolver.Address{Addr: string(data), ServerName: serviceName})
		}
	}
	return conns, nil
}

func (s *ZkClient) GetUserIdHashGatewayHost(ctx context.Context, userId string) (string, error) {
	glog.Slog.WarnKVs(ctx, "not implement", "err", "not implement")
	return "", nil
}

func (s *ZkClient) GetConns(ctx context.Context, serviceName string, opts ...grpc.DialOption) ([]*grpc.ClientConn, error) {
	glog.Slog.DebugKVs(GetZkCtx, "get conns from client", "serviceName", serviceName)
	s.lock.Lock()
	defer s.lock.Unlock()
	conns := s.localConns[serviceName]
	if len(conns) == 0 {
		glog.Slog.DebugKVs(GetZkCtx, "get conns from zk local", "serviceName", serviceName)
		addrs, err := s.GetConnsRemote(ctx, serviceName)
		if err != nil {
			return nil, err
		}
		if len(addrs) == 0 {
			return nil, gerr.New("addr is empty").WrapMsg("no conn for service", "serviceName",
				serviceName, "local conn", s.localConns, "ZkServers", s.ZkServers, "zkRoot", s.zkRoot)
		}
		for _, addr := range addrs {
			cc, err := grpc.DialContext(ctx, addr.Addr, append(s.options, opts...)...)
			if err != nil {
				glog.Slog.ErrorKVs(GetZkCtx, "dialContext failed", err, "addr", addr.Addr, "opts", append(s.options, opts...))
				return nil, gerr.WrapMsg(err, "DialContext failed", "addr.Addr", addr.Addr)
			}
			conns = append(conns, cc)
		}
		s.localConns[serviceName] = conns
	}
	return conns, nil
}

func (s *ZkClient) GetConn(ctx context.Context, serviceName string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	newOpts := append(s.options, grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, s.balancerName)))
	glog.Slog.DebugKVs(GetZkCtx, "get conn from client", "serviceName", serviceName, "opts", newOpts)
	return grpc.DialContext(ctx, fmt.Sprintf("%s:///%s", s.scheme, serviceName), append(newOpts, opts...)...)
}

func (s *ZkClient) GetSelfConnTarget() string {
	return s.rpcRegisterAddr
}

func (s *ZkClient) CloseConn(conn *grpc.ClientConn) {
	conn.Close()
}

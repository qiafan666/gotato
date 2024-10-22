package gzookeeper

import (
	"context"
	"github.com/qiafan666/gotato/commons/glog"
	"google.golang.org/grpc/resolver"
	"strings"
)

type Resolver struct {
	client         *ZkClient
	target         resolver.Target
	cc             resolver.ClientConn
	addrs          []resolver.Address
	getConnsRemote func(ctx context.Context, serviceName string) (conns []resolver.Address, err error)
}

func (r *Resolver) ResolveNowZK(o resolver.ResolveNowOptions) {
	serviceName := strings.TrimLeft(r.target.URL.Path, "/")
	glog.Slog.DebugKVs(GetZkCtx, "resolve now", "target", r.target, "serviceName", serviceName)
	newConns, err := r.getConnsRemote(context.Background(), serviceName)
	if err != nil {
		glog.Slog.ErrorKVs(context.Background(), "getConnsRemote error", err, "target", r.target, "serviceName", serviceName)
		glog.Slog.ErrorKVs(GetZkCtx, "getConnsRemote error", err, "target", r.target, "serviceName", serviceName)
		return
	}
	r.addrs = newConns
	if err := r.cc.UpdateState(resolver.State{Addresses: newConns}); err != nil {
		glog.Slog.ErrorKVs(GetZkCtx, "UpdateState error", err, "target", r.target, "serviceName", serviceName)
		return
	}
	glog.Slog.DebugKVs(GetZkCtx, "resolve now finished", "target", r.target, "conns", r.addrs, "serviceName", serviceName)
}

func (r *Resolver) ResolveNow(o resolver.ResolveNowOptions) {}

func (r *Resolver) Close() {}

func (s *ZkClient) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	glog.Slog.DebugKVs(GetZkCtx, "build resolver", "target", target, "cc", cc.UpdateState)
	serviceName := strings.TrimLeft(target.URL.Path, "/")
	r := &Resolver{client: s}
	r.target = target
	r.cc = cc
	r.getConnsRemote = s.GetConnsRemote
	r.ResolveNowZK(resolver.ResolveNowOptions{})
	s.lock.Lock()
	defer s.lock.Unlock()
	s.resolvers[serviceName] = r
	glog.Slog.DebugKVs(GetZkCtx, "build resolver finished", "target", target, "cc", cc.UpdateState, "key", serviceName)
	return r, nil
}

func (s *ZkClient) Scheme() string { return s.scheme }

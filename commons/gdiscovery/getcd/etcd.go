package getcd

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	gresolver "google.golang.org/grpc/resolver"
	"io"
	"strings"
	"sync"
	"time"
)

// ZkOption 定义了一个函数类型，用于修改 clientv3.Config
type ZkOption func(*clientv3.Config)

// SvcDiscoveryRegistryImpl 是服务发现和注册的实现
type SvcDiscoveryRegistryImpl struct {
	client            *clientv3.Client  // etcd 客户端
	resolver          gresolver.Builder // gRPC 解析器构建器
	dialOptions       []grpc.DialOption // gRPC 连接选项
	serviceKey        string            // 服务键
	endpointMgr       endpoints.Manager // 服务端点管理器
	leaseID           clientv3.LeaseID  // etcd 租约 ID
	rpcRegisterTarget string            // 注册的 RPC 目标

	rootDirectory string                        // etcd 根目录
	mu            sync.RWMutex                  // 读写锁保护连接映射
	connMap       map[string][]*grpc.ClientConn // gRPC 连接映射
}

// createNoOpLogger 创建一个无操作的日志记录器
func createNoOpLogger() *zap.Logger {
	noOpWriter := zapcore.AddSync(io.Discard)
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		noOpWriter,
		zapcore.InfoLevel,
	)
	return zap.New(core)
}

// NewSvcDiscoveryRegistry 创建一个新的服务发现注册实现
func NewSvcDiscoveryRegistry(rootDirectory string, endpoints []string, options ...ZkOption) (*SvcDiscoveryRegistryImpl, error) {
	cfg := clientv3.Config{
		Endpoints:           endpoints,
		DialTimeout:         5 * time.Second,
		PermitWithoutStream: true,
		Logger:              createNoOpLogger(),
		MaxCallSendMsgSize:  10 * 1024 * 1024,
	}

	// 应用传入的选项到配置中
	for _, opt := range options {
		opt(&cfg)
	}

	client, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}
	r, err := resolver.NewBuilder(client)
	if err != nil {
		return nil, err
	}

	s := &SvcDiscoveryRegistryImpl{
		client:        client,
		resolver:      r,
		rootDirectory: rootDirectory,
		connMap:       make(map[string][]*grpc.ClientConn),
	}

	go s.watchServiceChanges()
	return s, nil
}

// initializeConnMap 获取所有现有的端点并填充本地连接映射
func (r *SvcDiscoveryRegistryImpl) initializeConnMap() error {
	fullPrefix := fmt.Sprintf("%s/", r.rootDirectory)
	resp, err := r.client.Get(context.Background(), fullPrefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	r.connMap = make(map[string][]*grpc.ClientConn)
	for _, kv := range resp.Kvs {
		prefix, addr := r.splitEndpoint(string(kv.Key))
		conn, err := grpc.DialContext(context.Background(), addr, append(r.dialOptions, grpc.WithResolvers(r.resolver))...)
		if err != nil {
			continue
		}
		r.connMap[prefix] = append(r.connMap[prefix], conn)
	}
	return nil
}

// WithDialTimeout 设置自定义的拨号超时时间
func WithDialTimeout(timeout time.Duration) ZkOption {
	return func(cfg *clientv3.Config) {
		cfg.DialTimeout = timeout
	}
}

// WithMaxCallSendMsgSize 设置自定义的最大消息发送大小
func WithMaxCallSendMsgSize(size int) ZkOption {
	return func(cfg *clientv3.Config) {
		cfg.MaxCallSendMsgSize = size
	}
}

// WithUsernameAndPassword 设置 etcd 客户端的用户名和密码
func WithUsernameAndPassword(username, password string) ZkOption {
	return func(cfg *clientv3.Config) {
		cfg.Username = username
		cfg.Password = password
	}
}

// GetUserIdHashGatewayHost 返回指定用户 ID 哈希的网关主机（未实现）
func (r *SvcDiscoveryRegistryImpl) GetUserIdHashGatewayHost(ctx context.Context, userId string) (string, error) {
	return "", nil
}

// GetConns 返回指定服务名称的 gRPC 客户端连接列表
func (r *SvcDiscoveryRegistryImpl) GetConns(ctx context.Context, serviceName string, opts ...grpc.DialOption) ([]*grpc.ClientConn, error) {
	fullServiceKey := fmt.Sprintf("%s/%s", r.rootDirectory, serviceName)
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.connMap) == 0 {
		r.initializeConnMap()
	}
	return r.connMap[fullServiceKey], nil
}

// GetConn 返回指定服务名称的单个 gRPC 客户端连接
func (r *SvcDiscoveryRegistryImpl) GetConn(ctx context.Context, serviceName string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	target := fmt.Sprintf("etcd:///%s/%s", r.rootDirectory, serviceName)
	return grpc.DialContext(ctx, target, append(append(r.dialOptions, opts...), grpc.WithResolvers(r.resolver))...)
}

// GetSelfConnTarget 返回当前服务的连接目标
func (r *SvcDiscoveryRegistryImpl) GetSelfConnTarget() string {
	return r.rpcRegisterTarget
}

// AddOption 添加 gRPC 拨号选项到现有选项中
func (r *SvcDiscoveryRegistryImpl) AddOption(opts ...grpc.DialOption) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connMap = make(map[string][]*grpc.ClientConn)
	r.dialOptions = append(r.dialOptions, opts...)
}

// CloseConn 关闭指定的 gRPC 客户端连接
func (r *SvcDiscoveryRegistryImpl) CloseConn(conn *grpc.ClientConn) {
	conn.Close()
}

// Register 向 etcd 注册一个新的服务端点
func (r *SvcDiscoveryRegistryImpl) Register(serviceName, host string, port int, opts ...grpc.DialOption) error {
	r.serviceKey = fmt.Sprintf("%s/%s/%s:%d", r.rootDirectory, serviceName, host, port)
	em, err := endpoints.NewManager(r.client, r.rootDirectory+"/"+serviceName)
	if err != nil {
		return err
	}
	r.endpointMgr = em

	leaseResp, err := r.client.Grant(context.Background(), 30)
	if err != nil {
		return err
	}
	r.leaseID = leaseResp.ID

	r.rpcRegisterTarget = fmt.Sprintf("%s:%d", host, port)
	endpoint := endpoints.Endpoint{Addr: r.rpcRegisterTarget}

	err = em.AddEndpoint(context.TODO(), r.serviceKey, endpoint, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}

	go r.keepAliveLease(r.leaseID)
	return nil
}

// keepAliveLease 通过发送保活请求保持租约
func (r *SvcDiscoveryRegistryImpl) keepAliveLease(leaseID clientv3.LeaseID) {
	ch, err := r.client.KeepAlive(context.Background(), leaseID)
	if err != nil {
		return
	}
	for ka := range ch {
		if ka == nil {
			return
		}
	}
}

// watchServiceChanges 监视服务目录的变化
func (r *SvcDiscoveryRegistryImpl) watchServiceChanges() {
	watchChan := r.client.Watch(context.Background(), r.rootDirectory, clientv3.WithPrefix())
	for range watchChan {
		r.mu.RLock()
		r.initializeConnMap()
		r.mu.RUnlock()
	}
}

// refreshConnMap 获取最新的端点并更新本地连接映射
func (r *SvcDiscoveryRegistryImpl) refreshConnMap(prefix string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	fullPrefix := fmt.Sprintf("%s/", prefix)
	resp, err := r.client.Get(context.Background(), fullPrefix, clientv3.WithPrefix())
	if err != nil {
		return
	}
	r.connMap[prefix] = []*grpc.ClientConn{}
	for _, kv := range resp.Kvs {
		_, addr := r.splitEndpoint(string(kv.Key))
		conn, err := grpc.DialContext(context.Background(), addr, append(r.dialOptions, grpc.WithResolvers(r.resolver))...)
		if err != nil {
			continue
		}
		r.connMap[prefix] = append(r.connMap[prefix], conn)
	}
}

// splitEndpoint 拆分端点字符串为前缀和地址
func (r *SvcDiscoveryRegistryImpl) splitEndpoint(input string) (string, string) {
	lastSlashIndex := strings.LastIndex(input, "/")
	if lastSlashIndex != -1 {
		part1 := input[:lastSlashIndex]
		part2 := input[lastSlashIndex+1:]
		return part1, part2
	}
	return input, ""
}

// UnRegister 从 etcd 中移除服务端点
func (r *SvcDiscoveryRegistryImpl) UnRegister() error {
	if r.endpointMgr == nil {
		return fmt.Errorf("endpoint manager is not initialized")
	}
	err := r.endpointMgr.DeleteEndpoint(context.TODO(), r.serviceKey)
	if err != nil {
		return err
	}
	return nil
}

// Close 关闭 etcd 客户端连接
func (r *SvcDiscoveryRegistryImpl) Close() {
	if r.client != nil {
		_ = r.client.Close()
	}

	r.mu.Lock()
	defer r.mu.Unlock()
}

// Check 检查 etcd 是否正在运行，并根据需要创建根节点
func Check(ctx context.Context, etcdServers []string, etcdRoot string, createIfNotExist bool, options ...ZkOption) error {
	cfg := clientv3.Config{
		Endpoints: etcdServers,
	}
	for _, opt := range options {
		opt(&cfg)
	}
	client, err := clientv3.New(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to connect to etcd")
	}
	defer client.Close()

	var opCtx context.Context
	var cancel context.CancelFunc
	if cfg.DialTimeout != 0 {
		opCtx, cancel = context.WithTimeout(ctx, cfg.DialTimeout)
	} else {
		opCtx, cancel = context.WithTimeout(ctx, 10*time.Second)
	}
	defer cancel()

	resp, err := client.Get(opCtx, etcdRoot)
	if err != nil {
		return errors.Wrap(err, "failed to get the root node from etcd")
	}

	if len(resp.Kvs) == 0 {
		if createIfNotExist {
			var leaseTTL int64 = 10
			var leaseResp *clientv3.LeaseGrantResponse
			if leaseTTL > 0 {
				leaseResp, err = client.Grant(opCtx, leaseTTL)
				if err != nil {
					return errors.Wrap(err, "failed to create lease in etcd")
				}
			}
			putOpts := []clientv3.OpOption{}
			if leaseResp != nil {
				putOpts = append(putOpts, clientv3.WithLease(leaseResp.ID))
			}

			_, err := client.Put(opCtx, etcdRoot, "", putOpts...)
			if err != nil {
				return errors.Wrap(err, "failed to create the root node in etcd")
			}
		} else {
			return fmt.Errorf("root node %s does not exist in etcd", etcdRoot)
		}
	}
	return nil
}

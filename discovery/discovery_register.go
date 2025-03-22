package discovery

import (
	"context"
	"github.com/qiafan666/gotato/commons/gerr"
	"github.com/qiafan666/gotato/discovery/getcd"
	"github.com/qiafan666/gotato/discovery/gzookeeper"
	"google.golang.org/grpc"
	"time"
)

type Conn interface {
	GetConns(ctx context.Context, serviceName string, opts ...grpc.DialOption) ([]*grpc.ClientConn, error) //1
	GetConn(ctx context.Context, serviceName string, opts ...grpc.DialOption) (*grpc.ClientConn, error)    //2
	GetSelfConnTarget() string                                                                             //3
	AddOption(opts ...grpc.DialOption)                                                                     //4
	CloseConn(conn *grpc.ClientConn)                                                                       //5
}
type SvcDiscoveryRegistry interface {
	Conn
	Register(serviceName, host string, port int, opts ...grpc.DialOption) error //6
	UnRegister() error                                                          //7
	Close()
	GetUserIdHashGatewayHost(ctx context.Context, userId string) (string, error) //
}

type DiscoveryRegister struct {
	Zookeeper Zookeeper
	Etcd      Etcd
}
type Zookeeper struct {
	Schema   string   `json:"schema" yaml:"schema"` // zookeeper集群中项目区分
	Address  []string `json:"address" yaml:"address"`
	Username string   `json:"username" yaml:"username"`
	Password string   `json:"password" yaml:"password"`
}
type Etcd struct {
	RootDirectory string   `json:"root_directory" yaml:"root_directory"`
	Address       []string `json:"address" yaml:"address"`
	Username      string   `json:"username" yaml:"username"`
	Password      string   `json:"password" yaml:"password"`
}

// NewDiscoveryRegister 创建一个服务发现注册器
// mode: zookeeper or etcd
func NewDiscoveryRegister(mode string, discovery DiscoveryRegister) (SvcDiscoveryRegistry, error) {
	switch mode {
	case "zookeeper":
		return gzookeeper.NewZkClient(
			discovery.Zookeeper.Address,
			discovery.Zookeeper.Schema,
			gzookeeper.WithFreq(time.Hour),
			gzookeeper.WithUserNameAndPassword(discovery.Zookeeper.Username, discovery.Zookeeper.Password),
			gzookeeper.WithRoundRobin(),
			gzookeeper.WithTimeout(10),
		)
	case "etcd":
		return getcd.NewSvcDiscoveryRegistry(
			discovery.Etcd.RootDirectory,
			discovery.Etcd.Address,
			getcd.WithDialTimeout(10*time.Second),
			getcd.WithMaxCallSendMsgSize(20*1024*1024),
			getcd.WithUsernameAndPassword(discovery.Etcd.Username, discovery.Etcd.Password))
	default:
		return nil, gerr.New("unsupported discovery type", "mode", mode).Wrap()
	}
}

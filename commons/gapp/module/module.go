package module

import (
	"github.com/qiafan666/gotato/commons/gapp/chanrpc"
	"github.com/qiafan666/gotato/commons/gface"
)

// IModule 模块
type IModule interface {
	OnInit() error            // 初始化
	Run(closeSig chan bool)   // 启动
	OnDestroy()               // 销毁
	Name() string             // 名字
	ChanSrv() chanrpc.IServer // 消息通道
	Logger() gface.ILogger    // 日志
}

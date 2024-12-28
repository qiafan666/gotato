package module

import (
	"github.com/qiafan666/gotato/commons/gapp2/chanrpc"
)

// Module 模块
type Module interface {
	OnInit() error            // 初始化
	OnDestroy()               // 销毁
	Run(closeSig chan bool)   // 启动
	Name() string             // 名字
	ChanSrv() chanrpc.IServer // 消息通道
}

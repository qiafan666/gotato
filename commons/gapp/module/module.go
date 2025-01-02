package module

import (
	"github.com/qiafan666/gotato/commons/gapp/chanrpc"
	"github.com/qiafan666/gotato/commons/gface"
)

// Module 模块
type Module interface {
	OnInit() error            // 初始化
	OnDestroy()               // 销毁
	Run(closeSig chan bool)   // 启动
	Name() string             // 名字
	ChanSrv() chanrpc.IServer // 消息通道
	Logger() gface.Logger     // 日志
}

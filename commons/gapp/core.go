package gapp

import (
	"fmt"
	"github.com/qiafan666/gotato/commons/gapp/chanrpc"
	"github.com/qiafan666/gotato/commons/gapp/logger"
	"github.com/qiafan666/gotato/commons/gapp/module"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/gface"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"sync/atomic"
	"syscall"
)

// 节点全局状态
const (
	AppStateNone = iota // 未开始或已停止
	AppStateInit        // 正在初始化中
	AppStateRun         // 正在运行中
	AppStateStop        // 正在停止中
)

// 单例
var defaultApp = NewApp()

// mod 模块
type mod struct {
	mi       module.IModule
	closeSig chan bool
	wg       sync.WaitGroup
}

// DefaultApp 默认单例
func DefaultApp() *App {
	return defaultApp
}

// NewApp 创建App
func NewApp() *App {
	app := &App{
		closeSig: make(chan os.Signal, 1),
		state:    AppStateNone,
	}
	return app
}

// App 中的 modules 在初始化(通过 Start 或 Run) 之后不能变更
// App 有两种启停方式:
//  1. Start -> Stop: 手动启动和停止app，比较干净，通常用于测试代码
//  2. Run -> Terminate: 基于Start/Stop封装，自动监听OS Signal或通过Terminate来终止，通常用于真正的节点启动流程
type App struct {
	mods     []*mod
	state    int32
	closeSig chan os.Signal
	wg       sync.WaitGroup
}

// SetState 设置状态
func (app *App) setState(s int32) {
	atomic.StoreInt32(&app.state, s)
}

// GetState 获取状态
func (app *App) GetState() int32 {
	return atomic.LoadInt32(&app.state)
}

// Start 非阻塞启动app，需要在当前goroutine调用Stop来停止app
func (app *App) Start(l gface.ILogger, mods ...module.IModule) {

	if l == nil {
		panic("logger not initialized")
	}
	logger.DefaultLogger = l

	// 单个app不能启动两次
	if app.GetState() != AppStateNone {
		logger.DefaultLogger.ErrorF(nil, "app already started")
	}
	if len(mods) == 0 {
		return
	}
	app.wg.Add(1)
	// 注册module 并增加开关
	// register
	for _, mi := range mods {
		m := new(mod)
		m.mi = mi
		m.closeSig = make(chan bool, 1)
		app.mods = append(app.mods, m)
	}
	app.setState(AppStateInit)
	// 模块初始化
	for _, m := range app.mods {
		mi := m.mi
		if err := mi.OnInit(); err != nil {
			logger.DefaultLogger.ErrorF(nil, "module %v init error %v", reflect.TypeOf(mi), err)
		}
	}
	// 模块启动
	for _, m := range app.mods {
		m.wg.Add(1)
		go app.run(m)
	}
	app.setState(AppStateRun)
}

// Stop 停止App
func (app *App) Stop() {
	if app.GetState() == AppStateStop || app.GetState() == AppStateNone {
		return
	}
	app.setState(AppStateStop)
	// 先进后出
	for i := len(app.mods) - 1; i >= 0; i-- {
		m := app.mods[i]
		close(m.closeSig)
		m.wg.Wait()
		app.destroy(m)
	}
	app.mods = nil
	app.wg.Done()
	app.setState(AppStateNone)
}

// Stats 调试使用  简单查看各个 Service Chan 中的请求个数 用于压力测试性能监控
// goroutine safe
func (app *App) Stats() string {
	var ret string
	for _, m := range app.mods {
		if m.mi.ChanSrv() != nil {
			ret += fmt.Sprintf("chan: %v, len: %v \r\n", m.mi.Name(), m.mi.ChanSrv().Len())
		}
	}
	return ret
}

// GetChanSrv 获取指定名字模块的消息投递通道
// goroutine safe
func (app *App) GetChanSrv(name string) chanrpc.IServer {
	for _, m := range app.mods {
		if m.mi.Name() == name {
			return m.mi.ChanSrv()
		}
	}
	return nil
}

type IActorModule interface {
	GetActorChanSrv(actorID int64) chanrpc.IServer
}

// GetActorChanSrv 获取指定由Module管理的指定ActorID消息投递通道
func (app *App) GetActorChanSrv(name string, actorID int64) chanrpc.IServer {
	for _, m := range app.mods {
		if m.mi.Name() == name {
			am, ok := m.mi.(IActorModule)
			if !ok {
				return nil
			}
			return am.GetActorChanSrv(actorID)
		}
	}
	return nil
}

// run 启动所有模块
func (app *App) run(m *mod) {
	defer func() {
		stack := gcommon.PrintPanicStack()
		logger.DefaultLogger.ErrorF(nil, "app run panic error: %s", stack)
	}()
	defer m.wg.Done()
	m.mi.Run(m.closeSig)
}

// destroy 销毁模块
func (app *App) destroy(m *mod) {
	defer func() {
		stack := gcommon.PrintPanicStack()
		logger.DefaultLogger.ErrorF(nil, "app destroy panic error: %s", stack)
	}()
	m.mi.OnDestroy()
}

// Run 阻塞启动app，在监测到SIGINT SIGTERM信号时自动终止App
// 也可在任意goroutine调用Terminate来终止
func (app *App) Run(l gface.ILogger, mods ...module.IModule) {
	app.Start(l, mods...)
	// 信号监听 优雅退出
	for {
		signal.Notify(app.closeSig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		sig := <-app.closeSig
		logger.DefaultLogger.InfoF(nil, "server closing down (signal: %v)", sig)
		if sig == syscall.SIGHUP {
			continue
		}
		break
	}
	app.Stop()
}

// Terminate 用于模拟信号，终止Run，并等待app停止完成
// goroutine safe
func (app *App) Terminate() {
	if app.GetState() != AppStateRun {
		return
	}
	//app.closeSig <- syscall.SIGUSR1
	app.wg.Wait()
}

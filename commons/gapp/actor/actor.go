package actor

import (
	"github.com/qiafan666/gotato/commons/gapp/chanrpc"
	"github.com/qiafan666/gotato/commons/gapp/logger"
	"sync"
)

// Actor状态定义
const (
	StateNone = 0
	StateInit = 1
	StateRun  = 2
	StateStop = 3
)

type Creator func(int64) IActor

// IActor 逻辑层需要实现的接口
type IActor interface {
	OnInit(initData any) error
	Run(closeSig chan bool)
	OnDestroy()
	ChanSrv() chanrpc.IServer
	ChanCli() chanrpc.IClient
}

type Actor struct {
	id       int64  // 唯一id
	delegate IActor // 承载具体actor业务的goroutine
	state    int32  // 状态

	// 并发控制
	closeSig chan bool
	wg       sync.WaitGroup
}

func NewActor(id int64, delegate IActor) *Actor {
	return &Actor{
		id:       id,
		delegate: delegate,
		state:    StateNone,
		closeSig: make(chan bool),
	}
}

// InitAndRun Actor运行入口
func (a *Actor) InitAndRun(initData any, syncInitChan chan error) {
	a.state = StateInit
	err := a.delegate.OnInit(initData)
	syncInitChan <- err
	if err != nil {
		logger.DefaultLogger.DebugF(nil, "actor[%d] init failed: %v", a.id, err)
		return
	}
	a.state = StateRun
	a.delegate.Run(a.closeSig)
	a.state = StateStop
	a.delegate.OnDestroy()
}

// Stop 终止Actor
// syncWait == true 表示同步等待终止完成
func (a *Actor) Stop(syncWait bool) {
	logger.DefaultLogger.DebugF(nil, "actor Stop actor id begin:%v state:%v wait:%v", a.id, a.state, syncWait)
	a.closeSig <- true
	if syncWait {
		a.wg.Wait()
	}
	logger.DefaultLogger.DebugF(nil, "actor Stop actor id end:%v state:%v wait:%v", a.id, a.state, syncWait)
}

// Delegate 慎用（为了场景注入）
func (a *Actor) Delegate() IActor {
	return a.delegate
}

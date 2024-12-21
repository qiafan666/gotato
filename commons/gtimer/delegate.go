package gtimer

import "github.com/qiafan666/gotato/commons/iface"

// Delegate Timer代理
type Delegate struct {
	chanTimer  chan *Timer
	dispatcher *Dispatcher
}

// NewLogicDelegate 创建逻辑Timer代理:
// 1. 使用全局Logic Timer Dispatcher，因为可能收到外部逻辑时间偏移的影响
// 2. 维护自己的chanTimer，通常结合skeleton使用
// 该API通常用于逻辑层定时器，如科研/建造定时
func NewLogicDelegate() ITimerDelegate {
	return &Delegate{
		chanTimer:  make(chan *Timer, _timerOpChanSize),
		dispatcher: _logicDispatcher,
	}
}

// NewSysDelegate 创建系统Timer代理:
//  1. 使用全局System Timer Dispatcher，不受外部逻辑时间偏移的影响
//  2. 不维护自己的chanTimer, Timer的Callback直接在全局Dispatcher中执行
//     因此对该Callback的并发安全，非阻塞，以及执行效率都有更高的要求(否则会阻塞整个时间轮)
//
// 该API供tse框架使用，禁止逻辑层使用
func NewSysDelegate() ITimerAPI {
	return &Delegate{
		dispatcher: _sysDispatcher,
	}
}

// NewTimer 创建定时器
func (d *Delegate) NewTimer(timerType int32, timerID, timeout int64, cb timerCb) int64 {
	return d.dispatcher.NewTimer(timerType, timerID, timeout, cb, d.chanTimer)
}

// UpdateTimer 加速 Timer
func (d *Delegate) UpdateTimer(timerID, newEndTs int64) {
	d.dispatcher.UpdateTimer(timerID, newEndTs)
}

// BatchNewTimers 批量创建定时器
func (d *Delegate) BatchNewTimers(ops []*NewOp) []int64 {
	for _, op := range ops {
		op.OwnerChan = d.chanTimer
	}
	return d.dispatcher.BatchNewTimers(ops)
}

// CancelTimer 取消定时器
func (d *Delegate) CancelTimer(timerID int64) {
	d.dispatcher.CancelTimer(timerID)
}

// Logger 获取日志接口
func (d *Delegate) Logger() iface.Logger {
	return d.dispatcher.logger
}

// ChanTimer 仅对LogicDelegate有意义
func (d *Delegate) ChanTimer() chan *Timer {
	return d.chanTimer
}

// Exec 仅对LogicDelegate有意义
func (*Delegate) Exec(t *Timer) {
	t.Cb()
}

package timer

import (
	"github.com/qiafan666/gotato/commons/gface"
	"sync/atomic"
	"time"
)

var (
	// 逻辑层使用的定时器，使用外部注入的时间获取函数，可能受到外部逻辑调整时间的影响
	// 通常用于逻辑层定时器，如科研定时
	_logicDispatcher *Dispatcher
	// 系统层使用的定时器，使用系统真实时间获取函数
	// 通常用于框架定时器，如请求超时
	_sysDispatcher   *Dispatcher
	_timerOpChanSize = 1000
	_running         int32
)

// Run 启动定时器汞
// logicNowMs: 外部注入的获取时间接口，用于启动LogicDispatcher
//
//	并使LogicDispatcher支持时间调整的能力，如果传入nil，则使用真实系统时间
func Run(logicMowMs func() int64, logger gface.Logger) {
	success := atomic.CompareAndSwapInt32(&_running, 0, 1)
	if !success {
		logger.WarnF(nil, "timer dispatcher Run twice")
	}
	sysNowMs := func() int64 {
		return time.Now().UnixMilli()
	}
	_sysDispatcher = newDispatcher(sysNowMs, logger)
	_sysDispatcher.Run()

	if logicMowMs == nil {
		logicMowMs = sysNowMs
	}
	_logicDispatcher = newDispatcher(logicMowMs, logger)
	_logicDispatcher.Run()
}

// Stop 停止定时器汞
func Stop() {
	_logicDispatcher.Stop()
	_logicDispatcher = nil
	_sysDispatcher.Stop()
	_sysDispatcher = nil
	atomic.CompareAndSwapInt32(&_running, 1, 0)
}

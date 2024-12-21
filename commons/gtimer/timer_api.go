package gtimer

import "github.com/qiafan666/gotato/commons/iface"

// ITimerAPI 定时器
type ITimerAPI interface {
	// NewTimer 创建一个定时器
	NewTimer(typ int32, timerID, endTs int64, cb timerCb) int64
	// BatchNewTimers 批量创建定时器
	BatchNewTimers(ops []*NewOp) []int64
	// UpdateTimer 更新定时器截止时间
	UpdateTimer(timerID, endTs int64)
	// CancelTimer 取消定时器
	CancelTimer(timerID int64)
	// Logger 日志接口
	Logger() iface.Logger
}

// ITimerDelegate .
type ITimerDelegate interface {
	ITimerAPI
	ChanTimer() chan *Timer
	Exec(t *Timer)
}

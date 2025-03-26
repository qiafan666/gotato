package timer

import (
	"fmt"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/gface"
)

// Timer 默认定义
const (
	timerTick  = 8  // 最小粒度 ms 考虑将其定义为2^N 提高效率
	timerLevel = 24 // 时间段最大分级
)

// timer回调函数
type timerCb func(int64)

// 单个Timer
type Timer struct {
	typ   int32   // 做消息统计用
	id    int64   // ID
	endTs int64   // 到期时间戳 ms
	cb    timerCb // timer 回调
	// timer 触发后会写入该channel，由该channel的持有者执行回调
	// 如果ownerChan为nil，则直接在Dispatcher中执行回调
	ownerChan chan *Timer
	logger    gface.ILogger
}

// Cb 执行timer回调，在发起Timer的上下文中执行
func (t *Timer) Cb() {
	defer func() {
		if stack := gcommon.PrintPanicStack(); stack != "" {
			t.logger.ErrorF(nil, "Timer Cb panic error: %s", stack)
		}
	}()
	defer func() {
		t.cb = nil
	}()
	t.cb(t.id)
}

// GetStatName 获取消息统计Key
func (t *Timer) GetStatName() string {
	return fmt.Sprint("Timer_", t.typ)
}

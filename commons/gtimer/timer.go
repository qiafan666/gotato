package gtimer

import (
	"fmt"
	"github.com/qiafan666/gotato/commons/iface"
	"runtime"
	"unsafe"
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
	logger    iface.Logger
}

// Cb 执行timer回调，在发起Timer的上下文中执行
func (t *Timer) Cb() {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 2048)
			l := runtime.Stack(buf, false)
			b := buf[:l]
			stack := *(*string)(unsafe.Pointer(&b))
			t.logger.ErrorF("Timer Cb panic", "err", r, "stack", stack)
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

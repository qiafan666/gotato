package g

import (
	"github.com/qiafan666/gotato/commons/gapp/logger"
	"runtime"
	"unsafe"
)

// Go 非Groutine安全的异步任务
type Go struct {
	ChanCb    chan func()
	pendingGo int
}

// New 新建Go
func New(l int) *Go {
	if l <= 0 {
		return nil
	}
	g := new(Go)
	g.ChanCb = make(chan func(), l)
	return g
}

// SafeGo 执行异步任务
func (g *Go) SafeGo(f, cb func()) {
	g.pendingGo++
	go func() {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 2048)
				l := runtime.Stack(buf, false)
				b := buf[:l]
				stack := *(*string)(unsafe.Pointer(&b))
				logger.DefaultLogger.ErrorF("go SafeGo panic error : %v, stack : %s", r, stack)
			}
		}()
		defer func() {
			g.ChanCb <- cb
		}()
		f()
	}()
}

// Cb 执行回调
func (g *Go) Cb(cb func()) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 2048)
			l := runtime.Stack(buf, false)
			b := buf[:l]
			stack := *(*string)(unsafe.Pointer(&b))
			logger.DefaultLogger.ErrorF("go Cb panic error : %v, stack : %s", r, stack)
		}
	}()
	defer func() {
		g.pendingGo--
	}()
	if cb != nil {
		cb()
	}
}

// Close 等待所有异步任务执行结束
func (g *Go) Close() {
	for g.pendingGo > 0 {
		g.Cb(<-g.ChanCb)
	}
}

// Idle 是否闲置
func (g *Go) Idle() bool {
	return g.pendingGo == 0
}

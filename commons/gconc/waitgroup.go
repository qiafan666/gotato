package gconc

import (
	"github.com/qiafan666/gotato/commons/gconc/panics"
	"sync"
)

type IWaitGroup interface {
	Go(f func())
	Wait()
	WaitAndRecover() *panics.Recovered
}

func NewWaitGroup() IWaitGroup {
	return &waitGroup{}
}

type waitGroup struct {
	wg sync.WaitGroup
	pc panics.Catcher
}

func (h *waitGroup) Go(f func()) {
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		h.pc.Try(f)
	}()
}

func (h *waitGroup) Wait() {
	h.wg.Wait()

	h.pc.Repanic()
}

func (h *waitGroup) WaitAndRecover() *panics.Recovered {
	h.wg.Wait()

	return h.pc.Recovered()
}

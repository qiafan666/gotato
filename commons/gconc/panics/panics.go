package panics

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync/atomic"
)

type Catcher struct {
	recovered atomic.Pointer[Recovered]
}

func (p *Catcher) Try(f func()) {
	defer p.tryRecover()
	f()
}

func (p *Catcher) tryRecover() {
	if val := recover(); val != nil {
		rp := NewRecovered(1, val)
		p.recovered.CompareAndSwap(nil, &rp)
	}
}

func (p *Catcher) Repanic() {
	if val := p.Recovered(); val != nil {
		panic(val)
	}
}

func (p *Catcher) Recovered() *Recovered {
	return p.recovered.Load()
}

func NewRecovered(skip int, value any) Recovered {
	// 64 frames should be plenty
	var callers [64]uintptr
	n := runtime.Callers(skip+1, callers[:])
	return Recovered{
		Value:   value,
		Callers: callers[:n],
		Stack:   debug.Stack(),
	}
}

type Recovered struct {
	Value   any
	Callers []uintptr
	Stack   []byte
}

func (p *Recovered) String() string {
	return fmt.Sprintf("panic: %v\nstacktrace:\n%s\n", p.Value, p.Stack)
}

func (p *Recovered) AsError() error {
	if p == nil {
		return nil
	}
	return &ErrRecovered{*p}
}

type ErrRecovered struct{ Recovered }

var _ error = (*ErrRecovered)(nil)

func (p *ErrRecovered) Error() string { return p.String() }

func (p *ErrRecovered) Unwrap() error {
	if err, ok := p.Value.(error); ok {
		return err
	}
	return nil
}

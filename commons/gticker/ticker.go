package gticker

import (
	"context"
	"time"
)

type Ticker struct {
	duration time.Duration
	f        func()
}

func NewTicker(duration time.Duration, f func()) *Ticker {
	ticker := &Ticker{
		duration: duration,
	}
	ticker.addFunc(f)
	return ticker
}

// Func 定时执行方法
func (t *Ticker) addFunc(f func()) *Ticker {
	t.f = func() {
		f()
	}
	return t
}

func (t *Ticker) Run(ctx context.Context) {
	ticker := time.NewTimer(t.duration)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			t.f()
			ticker.Reset(t.duration)
		case <-ctx.Done():
			return
		}
	}
}

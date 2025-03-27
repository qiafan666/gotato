package stream

import (
	"github.com/qiafan666/gotato/commons/gconc"
	"github.com/qiafan666/gotato/commons/gconc/panics"
	"github.com/qiafan666/gotato/commons/gconc/pool"
	"sync"
)

func New() *Stream {
	return &Stream{
		pool: *pool.New(),
	}
}

type Stream struct {
	pool             pool.Pool
	callbackerHandle gconc.IWaitGroup
	queue            chan callbackCh

	initOnce sync.Once
}

type Task func() Callback

type Callback func()

func (s *Stream) Go(f Task) {
	s.init()

	ch := getCh()

	s.queue <- ch

	s.pool.Go(func() {
		defer func() {
			if r := recover(); r != nil {
				ch <- func() {}
				panic(r)
			}
		}()

		callback := f()
		ch <- callback
	})
}

func (s *Stream) Wait() {
	s.init()

	defer func() {
		close(s.queue)
		s.callbackerHandle.Wait()
	}()

	s.pool.Wait()
}

func (s *Stream) WithMaxGoroutines(n int) *Stream {
	s.pool.WithMaxGoroutines(n)
	return s
}

func (s *Stream) init() {
	s.initOnce.Do(func() {
		s.queue = make(chan callbackCh, s.pool.MaxGoroutines()+1)

		s.callbackerHandle.Go(s.callbacker)
	})
}

func (s *Stream) callbacker() {
	var panicCatcher panics.Catcher
	defer panicCatcher.Repanic()

	for callbackCh := range s.queue {
		callback := <-callbackCh

		if callback != nil {
			panicCatcher.Try(callback)
		}

		putCh(callbackCh)
	}
}

type callbackCh chan func()

var callbackChPool = sync.Pool{
	New: func() any {
		return make(callbackCh, 1)
	},
}

func getCh() callbackCh {
	return callbackChPool.Get().(callbackCh)
}

func putCh(ch callbackCh) {
	callbackChPool.Put(ch)
}

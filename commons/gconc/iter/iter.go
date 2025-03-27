package iter

import (
	"github.com/qiafan666/gotato/commons/gconc"
	"runtime"
	"sync/atomic"
)

func defaultMaxGoroutines() int { return runtime.GOMAXPROCS(0) }

type Iterator[T any] struct {
	MaxGoroutines int
}

func ForEach[T any](input []T, f func(*T)) { Iterator[T]{}.ForEach(input, f) }

func (iter Iterator[T]) ForEach(input []T, f func(*T)) {
	iter.ForEachIdx(input, func(_ int, t *T) {
		f(t)
	})
}

func ForEachIdx[T any](input []T, f func(int, *T)) { Iterator[T]{}.ForEachIdx(input, f) }

func (iter Iterator[T]) ForEachIdx(input []T, f func(int, *T)) {
	if iter.MaxGoroutines == 0 {
		iter.MaxGoroutines = defaultMaxGoroutines()
	}

	numInput := len(input)
	if iter.MaxGoroutines > numInput {
		iter.MaxGoroutines = numInput
	}

	var idx atomic.Int64
	task := func() {
		i := int(idx.Add(1) - 1)
		for ; i < numInput; i = int(idx.Add(1) - 1) {
			f(i, &input[i])
		}
	}

	var wg gconc.IWaitGroup
	for i := 0; i < iter.MaxGoroutines; i++ {
		wg.Go(task)
	}
	wg.Wait()
}

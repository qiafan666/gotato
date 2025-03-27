package pool

import (
	"context"
	"sort"
	"sync"
)

func NewWithResults[T any]() *ResultPool[T] {
	return &ResultPool[T]{
		pool: *New(),
	}
}

type ResultPool[T any] struct {
	pool Pool
	agg  resultAggregator[T]
}

func (p *ResultPool[T]) Go(f func() T) {
	idx := p.agg.nextIndex()
	p.pool.Go(func() {
		p.agg.save(idx, f(), false)
	})
}

func (p *ResultPool[T]) Wait() []T {
	p.pool.Wait()
	results := p.agg.collect(true)
	p.agg = resultAggregator[T]{} // reset for reuse
	return results
}

func (p *ResultPool[T]) MaxGoroutines() int {
	return p.pool.MaxGoroutines()
}

func (p *ResultPool[T]) WithErrors() *ResultErrorPool[T] {
	p.panicIfInitialized()
	return &ResultErrorPool[T]{
		errorPool: *p.pool.WithErrors(),
	}
}

func (p *ResultPool[T]) WithContext(ctx context.Context) *ResultContextPool[T] {
	p.panicIfInitialized()
	return &ResultContextPool[T]{
		contextPool: *p.pool.WithContext(ctx),
	}
}

func (p *ResultPool[T]) WithMaxGoroutines(n int) *ResultPool[T] {
	p.panicIfInitialized()
	p.pool.WithMaxGoroutines(n)
	return p
}

func (p *ResultPool[T]) panicIfInitialized() {
	p.pool.panicIfInitialized()
}

type resultAggregator[T any] struct {
	mu      sync.Mutex
	len     int
	results []T
	errored []int
}

func (r *resultAggregator[T]) nextIndex() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	nextIdx := r.len
	r.len += 1
	return nextIdx
}

func (r *resultAggregator[T]) save(i int, res T, errored bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if i >= len(r.results) {
		old := r.results
		r.results = make([]T, r.len)
		copy(r.results, old)
	}

	r.results[i] = res

	if errored {
		r.errored = append(r.errored, i)
	}
}

func (r *resultAggregator[T]) collect(collectErrored bool) []T {
	if !r.mu.TryLock() {
		panic("collect should not be called until all goroutines have exited")
	}

	if collectErrored || len(r.errored) == 0 {
		return r.results
	}

	filtered := r.results[:0]
	sort.Ints(r.errored)
	for i, e := range r.errored {
		if i == 0 {
			filtered = append(filtered, r.results[:e]...)
		} else {
			filtered = append(filtered, r.results[r.errored[i-1]+1:e]...)
		}
	}
	return filtered
}

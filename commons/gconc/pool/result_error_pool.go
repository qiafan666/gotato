package pool

import (
	"context"
)

type ResultErrorPool[T any] struct {
	errorPool      ErrorPool
	agg            resultAggregator[T]
	collectErrored bool
}

func (p *ResultErrorPool[T]) Go(f func() (T, error)) {
	idx := p.agg.nextIndex()
	p.errorPool.Go(func() error {
		res, err := f()
		p.agg.save(idx, res, err != nil)
		return err
	})
}

func (p *ResultErrorPool[T]) Wait() ([]T, error) {
	err := p.errorPool.Wait()
	results := p.agg.collect(p.collectErrored)
	p.agg = resultAggregator[T]{} // reset for reuse
	return results, err
}

func (p *ResultErrorPool[T]) WithCollectErrored() *ResultErrorPool[T] {
	p.panicIfInitialized()
	p.collectErrored = true
	return p
}

func (p *ResultErrorPool[T]) WithContext(ctx context.Context) *ResultContextPool[T] {
	p.panicIfInitialized()
	return &ResultContextPool[T]{
		contextPool: *p.errorPool.WithContext(ctx),
	}
}

func (p *ResultErrorPool[T]) WithFirstError() *ResultErrorPool[T] {
	p.panicIfInitialized()
	p.errorPool.WithFirstError()
	return p
}

func (p *ResultErrorPool[T]) WithMaxGoroutines(n int) *ResultErrorPool[T] {
	p.panicIfInitialized()
	p.errorPool.WithMaxGoroutines(n)
	return p
}

func (p *ResultErrorPool[T]) panicIfInitialized() {
	p.errorPool.panicIfInitialized()
}

package pool

import (
	"context"
)

type ResultContextPool[T any] struct {
	contextPool    ContextPool
	agg            resultAggregator[T]
	collectErrored bool
}

func (p *ResultContextPool[T]) Go(f func(context.Context) (T, error)) {
	idx := p.agg.nextIndex()
	p.contextPool.Go(func(ctx context.Context) error {
		res, err := f(ctx)
		p.agg.save(idx, res, err != nil)
		return err
	})
}

func (p *ResultContextPool[T]) Wait() ([]T, error) {
	err := p.contextPool.Wait()
	results := p.agg.collect(p.collectErrored)
	p.agg = resultAggregator[T]{}
	return results, err
}

func (p *ResultContextPool[T]) WithCollectErrored() *ResultContextPool[T] {
	p.panicIfInitialized()
	p.collectErrored = true
	return p
}

func (p *ResultContextPool[T]) WithFirstError() *ResultContextPool[T] {
	p.panicIfInitialized()
	p.contextPool.WithFirstError()
	return p
}

func (p *ResultContextPool[T]) WithCancelOnError() *ResultContextPool[T] {
	p.panicIfInitialized()
	p.contextPool.WithCancelOnError()
	return p
}

func (p *ResultContextPool[T]) WithFailFast() *ResultContextPool[T] {
	p.panicIfInitialized()
	p.contextPool.WithFailFast()
	return p
}

func (p *ResultContextPool[T]) WithMaxGoroutines(n int) *ResultContextPool[T] {
	p.panicIfInitialized()
	p.contextPool.WithMaxGoroutines(n)
	return p
}

func (p *ResultContextPool[T]) panicIfInitialized() {
	p.contextPool.panicIfInitialized()
}

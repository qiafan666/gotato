package pool

import (
	"context"
)

type ContextPool struct {
	errorPool ErrorPool

	ctx    context.Context
	cancel context.CancelFunc

	cancelOnError bool
}

func (p *ContextPool) Go(f func(ctx context.Context) error) {
	p.errorPool.Go(func() error {
		if p.cancelOnError {
			defer func() {
				if r := recover(); r != nil {
					p.cancel()
					panic(r)
				}
			}()
		}

		err := f(p.ctx)
		if err != nil && p.cancelOnError {
			p.errorPool.addErr(err)
			p.cancel()
			return nil
		}
		return err
	})
}

func (p *ContextPool) Wait() error {
	defer p.cancel()
	return p.errorPool.Wait()
}

func (p *ContextPool) WithFirstError() *ContextPool {
	p.panicIfInitialized()
	p.errorPool.WithFirstError()
	return p
}

func (p *ContextPool) WithCancelOnError() *ContextPool {
	p.panicIfInitialized()
	p.cancelOnError = true
	return p
}

func (p *ContextPool) WithFailFast() *ContextPool {
	p.panicIfInitialized()
	p.WithFirstError()
	p.WithCancelOnError()
	return p
}

func (p *ContextPool) WithMaxGoroutines(n int) *ContextPool {
	p.panicIfInitialized()
	p.errorPool.WithMaxGoroutines(n)
	return p
}

func (p *ContextPool) panicIfInitialized() {
	p.errorPool.panicIfInitialized()
}

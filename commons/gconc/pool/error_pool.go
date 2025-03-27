package pool

import (
	"context"
	"errors"
	"sync"
)

type ErrorPool struct {
	pool Pool

	onlyFirstError bool

	mu   sync.Mutex
	errs []error
}

func (p *ErrorPool) Go(f func() error) {
	p.pool.Go(func() {
		p.addErr(f())
	})
}

func (p *ErrorPool) Wait() error {
	p.pool.Wait()

	errs := p.errs
	p.errs = nil // reset errs

	if len(errs) == 0 {
		return nil
	} else if p.onlyFirstError {
		return errs[0]
	} else {
		return errors.Join(errs...)
	}
}

func (p *ErrorPool) WithContext(ctx context.Context) *ContextPool {
	p.panicIfInitialized()
	ctx, cancel := context.WithCancel(ctx)
	return &ContextPool{
		errorPool: p.deref(),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (p *ErrorPool) WithFirstError() *ErrorPool {
	p.panicIfInitialized()
	p.onlyFirstError = true
	return p
}

func (p *ErrorPool) WithMaxGoroutines(n int) *ErrorPool {
	p.panicIfInitialized()
	p.pool.WithMaxGoroutines(n)
	return p
}

func (p *ErrorPool) deref() ErrorPool {
	return ErrorPool{
		pool:           p.pool.deref(),
		onlyFirstError: p.onlyFirstError,
	}
}

func (p *ErrorPool) panicIfInitialized() {
	p.pool.panicIfInitialized()
}

func (p *ErrorPool) addErr(err error) {
	if err != nil {
		p.mu.Lock()
		p.errs = append(p.errs, err)
		p.mu.Unlock()
	}
}

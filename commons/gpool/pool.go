package gpool

import (
	"context"
	"errors"
	"io"
	"reflect"
	"sync"
)

func IsZero[T any](v T) bool {
	return reflect.ValueOf(v).IsZero()
}

type Item interface {
	io.Closer
	IsClosed() bool
	CloseNotify() <-chan any
}

var (
	ErrInvalidOptions = errors.New("invalid options")
	ErrClosed         = errors.New("pool closed")
	ErrCancel         = errors.New("operation canceled")
)

type Options[T Item] struct {
	MaxSize  uint
	InitSize uint
	New      func() (T, error)
}

type Pool[T Item] struct {
	ctx context.Context

	items chan T

	mu        sync.Mutex
	closeOnce sync.Once

	size uint

	opt *Options[T]
}

func NewPool[T Item](ctx context.Context, opt *Options[T]) (*Pool[T], error) {
	p := &Pool[T]{
		ctx:   ctx,
		items: nil,
		opt:   opt,
	}

	if p.opt.MaxSize == 0 {
		return nil, ErrInvalidOptions
	}
	if p.opt.InitSize > p.opt.MaxSize {
		return nil, ErrInvalidOptions
	}

	p.items = make(chan T, p.opt.MaxSize)

	if p.opt.InitSize == 0 {
		return p, nil
	}
	for i := 0; i < int(p.opt.InitSize); i++ {
		v, err := p.newItem()
		if err != nil {
			return p, err
		}
		p.items <- v
	}

	return p, nil
}

func (p *Pool[T]) Get(ctx context.Context) (T, error) {
	if ctx.Err() != nil {
		var t T
		return t, ErrClosed
	}
	v, err := p.getItemInstance(ctx)
	if err != nil {
		return v, err
	}
	if !v.IsClosed() {
		return v, nil
	}

	v.Close()
	return p.Get(ctx)
}

func (p *Pool[T]) getItemInstance(ctx context.Context) (T, error) {
	var v T
	if p.items == nil {
		return v, ErrClosed
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.size >= p.opt.MaxSize {
		select {
		case <-ctx.Done():
			return v, ErrCancel
		case v := <-p.items:
			if IsZero[T](v) {
				return v, ErrClosed
			}
			return v, nil
		}
	}

	select {
	case v := <-p.items:
		if IsZero[T](v) {
			return v, ErrClosed
		}
		return v, nil
	default:
		return p.newItem()
	}
}

func (p *Pool[T]) newItem() (T, error) {
	v, err := p.opt.New()
	if err != nil {
		if !IsZero[T](v) {
			v.Close()
		}
		return v, err
	}

	p.size++

	go func() {
		select {
		case <-v.CloseNotify():
			p.closeItem(v)
		case <-p.ctx.Done():
			return
		}
	}()
	return v, nil
}

func (p *Pool[T]) Put(v T) {
	if IsZero[T](v) {
		return
	}

	if p.items == nil || v.IsClosed() {
		v.Close()
		return
	}

	select {
	case p.items <- v:
		return
	default:
		// pool已满,丢弃
		v.Close()
		return
	}
}

func (p *Pool[T]) Close() {
	p.closeOnce.Do(func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		items := p.items
		p.items = nil
		p.opt = nil

		if items == nil {
			return
		}

		close(items)
		for v := range items {
			v.Close()
		}
	})
}

func (p *Pool[T]) closeItem(v T) {
	if IsZero[T](v) {
		return
	}
	v.Close()
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.size > 0 {
		p.size--
	}
}

package gcommon

import "sync"

type Chan[T any] struct {
	ch chan T

	closeOnce sync.Once
	closed    bool
}

func NewChan[T any](bufSize uint) *Chan[T] {
	c := &Chan[T]{
		ch:        make(chan T, bufSize),
		closeOnce: sync.Once{},
		closed:    false,
	}

	return c
}

func (c *Chan[T]) Write(p T) {
	if c.closed {
		return
	}
	c.ch <- p
}

func (c *Chan[T]) Read() <-chan T {
	return c.ch
}

func (c *Chan[T]) Close() {
	c.closeOnce.Do(func() {
		c.closed = true
		close(c.ch)
	})
}

func (c *Chan[T]) IsClosed() bool {
	return c.closed
}

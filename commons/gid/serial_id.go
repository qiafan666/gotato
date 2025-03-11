package gid

import (
	"fmt"
	"sync"
)

type t interface {
	comparable
	uint | uint8 | uint16 | uint32 | uint64
}
type SerialId[T t] struct {
	mu sync.Mutex
	id T
}

func NewSerialId[T t]() *SerialId[T] {
	return new(SerialId[T])
}

func (n *SerialId[T]) Id() T {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.id++
	return n.id
}

func (n *SerialId[T]) StringId() string {
	return fmt.Sprintf("%d", n.Id())
}

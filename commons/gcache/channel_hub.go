package gcache

import (
	"sync"
)

type ChannelHub[T any] struct {
	mu       sync.RWMutex
	channels map[string]chan T
	bufSize  int
}

func NewChannelHub[T any](bufSize int) *ChannelHub[T] {
	return &ChannelHub[T]{
		channels: make(map[string]chan T),
		bufSize:  bufSize,
	}
}

// Subscribe 订阅某个 channel，返回一个只读 channel
func (h *ChannelHub[T]) Subscribe(symbol string) <-chan T {
	h.mu.Lock()
	defer h.mu.Unlock()

	if ch, ok := h.channels[symbol]; ok {
		return ch
	}
	ch := make(chan T, h.bufSize)
	h.channels[symbol] = ch
	return ch
}

// Publish 往某个 channel 推消息
func (h *ChannelHub[T]) Publish(symbol string, msg T) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if ch, ok := h.channels[symbol]; ok {
		select {
		case ch <- msg:
		default:
			// 如果满了，可以选择丢弃/覆盖/阻塞，这里先丢弃
		}
	}
}

// CloseAll 关闭所有 channel
func (h *ChannelHub[T]) CloseAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for sym, ch := range h.channels {
		close(ch)
		delete(h.channels, sym)
	}
}

// CloseChannel 关闭某个 channel
func (h *ChannelHub[T]) CloseChannel(symbol string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if ch, ok := h.channels[symbol]; ok {
		close(ch)
		delete(h.channels, symbol)
	}
}

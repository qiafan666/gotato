package simplelru

import (
	"errors"
	"github.com/qiafan666/gotato/commons/gcache/glru/internal"
)

type EvictCallback[K comparable, V any] func(key K, value V)

type LRU[K comparable, V any] struct {
	size      int
	evictList *internal.LruList[K, V]
	items     map[K]*internal.Entry[K, V]
	onEvict   EvictCallback[K, V]
}

func NewLRU[K comparable, V any](size int, onEvict EvictCallback[K, V]) (*LRU[K, V], error) {
	if size <= 0 {
		return nil, errors.New("must provide a positive size")
	}

	c := &LRU[K, V]{
		size:      size,
		evictList: internal.NewList[K, V](),
		items:     make(map[K]*internal.Entry[K, V]),
		onEvict:   onEvict,
	}
	return c, nil
}

func (c *LRU[K, V]) Purge() {
	for k, v := range c.items {
		if c.onEvict != nil {
			c.onEvict(k, v.Value)
		}
		delete(c.items, k)
	}
	c.evictList.Init()
}

func (c *LRU[K, V]) Add(key K, value V) (evicted bool) {
	// Check for existing item
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		ent.Value = value
		return false
	}

	// Add new item
	ent := c.evictList.PushFront(key, value)
	c.items[key] = ent

	evict := c.evictList.Length() > c.size
	// Verify size not exceeded
	if evict {
		c.removeOldest()
	}
	return evict
}

func (c *LRU[K, V]) Get(key K) (value V, ok bool) {
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		return ent.Value, true
	}
	return
}

func (c *LRU[K, V]) Contains(key K) (ok bool) {
	_, ok = c.items[key]
	return ok
}

func (c *LRU[K, V]) Peek(key K) (value V, ok bool) {
	var ent *internal.Entry[K, V]
	if ent, ok = c.items[key]; ok {
		return ent.Value, true
	}
	return
}

func (c *LRU[K, V]) Remove(key K) (present bool) {
	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
		return true
	}
	return false
}

func (c *LRU[K, V]) RemoveOldest() (key K, value V, ok bool) {
	if ent := c.evictList.Back(); ent != nil {
		c.removeElement(ent)
		return ent.Key, ent.Value, true
	}
	return
}

func (c *LRU[K, V]) GetOldest() (key K, value V, ok bool) {
	if ent := c.evictList.Back(); ent != nil {
		return ent.Key, ent.Value, true
	}
	return
}

func (c *LRU[K, V]) Keys() []K {
	keys := make([]K, c.evictList.Length())
	i := 0
	for ent := c.evictList.Back(); ent != nil; ent = ent.PrevEntry() {
		keys[i] = ent.Key
		i++
	}
	return keys
}

func (c *LRU[K, V]) Values() []V {
	values := make([]V, len(c.items))
	i := 0
	for ent := c.evictList.Back(); ent != nil; ent = ent.PrevEntry() {
		values[i] = ent.Value
		i++
	}
	return values
}

func (c *LRU[K, V]) Len() int {
	return c.evictList.Length()
}

func (c *LRU[K, V]) Cap() int {
	return c.size
}

func (c *LRU[K, V]) Resize(size int) (evicted int) {
	diff := c.Len() - size
	if diff < 0 {
		diff = 0
	}
	for i := 0; i < diff; i++ {
		c.removeOldest()
	}
	c.size = size
	return diff
}

func (c *LRU[K, V]) removeOldest() {
	if ent := c.evictList.Back(); ent != nil {
		c.removeElement(ent)
	}
}

func (c *LRU[K, V]) removeElement(e *internal.Entry[K, V]) {
	c.evictList.Remove(e)
	delete(c.items, e.Key)
	if c.onEvict != nil {
		c.onEvict(e.Key, e.Value)
	}
}

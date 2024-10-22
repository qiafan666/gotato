package arc

import (
	"github.com/qiafan666/gotato/commons/gcache/glru/simplelru"
	"sync"
)

// ARCCache 是一个线程安全的固定大小自适应替换缓存 (ARC)。
// ARC 是对标准 LRU 缓存的增强，因为它同时跟踪使用频率和最近使用情况。
// 这避免了对新条目的突发访问会驱逐掉那些频繁使用的较旧条目。
// 与标准的 LRU 缓存相比，它增加了一些额外的跟踪开销，从计算角度来看，
// 其成本大约是 LRU 的 2 倍，并且额外的内存开销与缓存大小呈线性关系。
type ARCCache[K comparable, V any] struct {
	size int // Size is the total capacity of the cache
	p    int // P is the dynamic preference towards T1 or T2

	t1 simplelru.LRUCache[K, V]        // T1 is the LRU for recently accessed items
	b1 simplelru.LRUCache[K, struct{}] // B1 is the LRU for evictions from t1

	t2 simplelru.LRUCache[K, V]        // T2 is the LRU for frequently accessed items
	b2 simplelru.LRUCache[K, struct{}] // B2 is the LRU for evictions from t2

	lock sync.RWMutex
}

// NewARC creates an ARC of the given size
func NewARC[K comparable, V any](size int) (*ARCCache[K, V], error) {
	// Create the sub LRUs
	b1, err := simplelru.NewLRU[K, struct{}](size, nil)
	if err != nil {
		return nil, err
	}
	b2, err := simplelru.NewLRU[K, struct{}](size, nil)
	if err != nil {
		return nil, err
	}
	t1, err := simplelru.NewLRU[K, V](size, nil)
	if err != nil {
		return nil, err
	}
	t2, err := simplelru.NewLRU[K, V](size, nil)
	if err != nil {
		return nil, err
	}

	// Initialize the ARC
	c := &ARCCache[K, V]{
		size: size,
		p:    0,
		t1:   t1,
		b1:   b1,
		t2:   t2,
		b2:   b2,
	}
	return c, nil
}

// Get looks up a key's value from the cache.
func (c *ARCCache[K, V]) Get(key K) (value V, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// If the value is contained in T1 (recent), then
	// promote it to T2 (frequent)
	if val, ok := c.t1.Peek(key); ok {
		c.t1.Remove(key)
		c.t2.Add(key, val)
		return val, ok
	}

	// Check if the value is contained in T2 (frequent)
	if val, ok := c.t2.Get(key); ok {
		return val, ok
	}

	// No hit
	return
}

// Add adds a value to the cache.
func (c *ARCCache[K, V]) Add(key K, value V) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Check if the value is contained in T1 (recent), and potentially
	// promote it to frequent T2
	if c.t1.Contains(key) {
		c.t1.Remove(key)
		c.t2.Add(key, value)
		return
	}

	// Check if the value is already in T2 (frequent) and update it
	if c.t2.Contains(key) {
		c.t2.Add(key, value)
		return
	}

	// Check if this value was recently evicted as part of the
	// recently used list
	if c.b1.Contains(key) {
		// T1 set is too small, increase P appropriately
		delta := 1
		b1Len := c.b1.Len()
		b2Len := c.b2.Len()
		if b2Len > b1Len {
			delta = b2Len / b1Len
		}
		if c.p+delta >= c.size {
			c.p = c.size
		} else {
			c.p += delta
		}

		// Potentially need to make room in the cache
		if c.t1.Len()+c.t2.Len() >= c.size {
			c.replace(false)
		}

		// Remove from B1
		c.b1.Remove(key)

		// Add the key to the frequently used list
		c.t2.Add(key, value)
		return
	}

	// Check if this value was recently evicted as part of the
	// frequently used list
	if c.b2.Contains(key) {
		// T2 set is too small, decrease P appropriately
		delta := 1
		b1Len := c.b1.Len()
		b2Len := c.b2.Len()
		if b1Len > b2Len {
			delta = b1Len / b2Len
		}
		if delta >= c.p {
			c.p = 0
		} else {
			c.p -= delta
		}

		// Potentially need to make room in the cache
		if c.t1.Len()+c.t2.Len() >= c.size {
			c.replace(true)
		}

		// Remove from B2
		c.b2.Remove(key)

		// Add the key to the frequently used list
		c.t2.Add(key, value)
		return
	}

	// Potentially need to make room in the cache
	if c.t1.Len()+c.t2.Len() >= c.size {
		c.replace(false)
	}

	// Keep the size of the ghost buffers trim
	if c.b1.Len() > c.size-c.p {
		c.b1.RemoveOldest()
	}
	if c.b2.Len() > c.p {
		c.b2.RemoveOldest()
	}

	// Add to the recently seen list
	c.t1.Add(key, value)
}

// replace is used to adaptively evict from either T1 or T2
// based on the current learned value of P
func (c *ARCCache[K, V]) replace(b2ContainsKey bool) {
	t1Len := c.t1.Len()
	if t1Len > 0 && (t1Len > c.p || (t1Len == c.p && b2ContainsKey)) {
		k, _, ok := c.t1.RemoveOldest()
		if ok {
			c.b1.Add(k, struct{}{})
		}
	} else {
		k, _, ok := c.t2.RemoveOldest()
		if ok {
			c.b2.Add(k, struct{}{})
		}
	}
}

// Len returns the number of cached entries
func (c *ARCCache[K, V]) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.t1.Len() + c.t2.Len()
}

// Cap returns the capacity of the cache
func (c *ARCCache[K, V]) Cap() int {
	return c.size
}

// Keys returns all the cached keys
func (c *ARCCache[K, V]) Keys() []K {
	c.lock.RLock()
	defer c.lock.RUnlock()
	k1 := c.t1.Keys()
	k2 := c.t2.Keys()
	return append(k1, k2...)
}

// Values returns all the cached values
func (c *ARCCache[K, V]) Values() []V {
	c.lock.RLock()
	defer c.lock.RUnlock()
	v1 := c.t1.Values()
	v2 := c.t2.Values()
	return append(v1, v2...)
}

// Remove is used to purge a key from the cache
func (c *ARCCache[K, V]) Remove(key K) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.t1.Remove(key) {
		return
	}
	if c.t2.Remove(key) {
		return
	}
	if c.b1.Remove(key) {
		return
	}
	if c.b2.Remove(key) {
		return
	}
}

// Purge is used to clear the cache
func (c *ARCCache[K, V]) Purge() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.t1.Purge()
	c.t2.Purge()
	c.b1.Purge()
	c.b2.Purge()
}

// Contains is used to check if the cache contains a key
// without updating recency or frequency.
func (c *ARCCache[K, V]) Contains(key K) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.t1.Contains(key) || c.t2.Contains(key)
}

// Peek is used to inspect the cache value of a key
// without updating recency or frequency.
func (c *ARCCache[K, V]) Peek(key K) (value V, ok bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if val, ok := c.t1.Peek(key); ok {
		return val, ok
	}
	return c.t2.Peek(key)
}

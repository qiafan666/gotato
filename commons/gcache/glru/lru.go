package glru

import (
	"github.com/qiafan666/gotato/commons/gcache/glru/simplelru"
	"sync"
)

const (
	// DefaultEvictedBufferSize 定义了存储被逐出键/值的默认缓冲区大小
	DefaultEvictedBufferSize = 16
)

// Cache 是一个线程安全的固定大小 LRU 缓存
type Cache[K comparable, V any] struct {
	lru         *simplelru.LRU[K, V] // LRU 缓存实例
	evictedKeys []K                  // 被逐出的键列表
	evictedVals []V                  // 被逐出的值列表
	onEvictedCB func(k K, v V)       // 键值对被逐出时的回调函数
	lock        sync.RWMutex         // 读写锁，保证线程安全
}

// New 创建一个指定大小的 LRU 缓存
func New[K comparable, V any](size int) (*Cache[K, V], error) {
	return NewWithEvict[K, V](size, nil)
}

// NewWithEvict 构造一个固定大小的缓存，并指定逐出回调函数
func NewWithEvict[K comparable, V any](size int, onEvicted func(key K, value V)) (c *Cache[K, V], err error) {
	// 创建一个带有默认设置的缓存
	c = &Cache[K, V]{
		onEvictedCB: onEvicted,
	}
	if onEvicted != nil {
		c.initEvictBuffers()
		onEvicted = c.onEvicted
	}
	c.lru, err = simplelru.NewLRU(size, onEvicted)
	return
}

// 初始化缓存的逐出缓冲区
func (c *Cache[K, V]) initEvictBuffers() {
	c.evictedKeys = make([]K, 0, DefaultEvictedBufferSize)
	c.evictedVals = make([]V, 0, DefaultEvictedBufferSize)
}

// onEvicted 保存被逐出的键/值对，并在临界区之外调用注册的回调函数
func (c *Cache[K, V]) onEvicted(k K, v V) {
	c.evictedKeys = append(c.evictedKeys, k)
	c.evictedVals = append(c.evictedVals, v)
}

// Purge 用于完全清空缓存
func (c *Cache[K, V]) Purge() {
	var ks []K
	var vs []V
	c.lock.Lock()
	c.lru.Purge()
	if c.onEvictedCB != nil && len(c.evictedKeys) > 0 {
		ks, vs = c.evictedKeys, c.evictedVals
		c.initEvictBuffers()
	}
	c.lock.Unlock()
	// 在临界区之外调用回调函数
	if c.onEvictedCB != nil {
		for i := 0; i < len(ks); i++ {
			c.onEvictedCB(ks[i], vs[i])
		}
	}
}

// Add 向缓存添加一个值。如果发生了逐出操作，返回 true
func (c *Cache[K, V]) Add(key K, value V) (evicted bool) {
	var k K
	var v V
	c.lock.Lock()
	evicted = c.lru.Add(key, value)
	if c.onEvictedCB != nil && evicted {
		k, v = c.evictedKeys[0], c.evictedVals[0]
		c.evictedKeys, c.evictedVals = c.evictedKeys[:0], c.evictedVals[:0]
	}
	c.lock.Unlock()
	if c.onEvictedCB != nil && evicted {
		c.onEvictedCB(k, v)
	}
	return
}

// Get 从缓存中查找一个键的值
func (c *Cache[K, V]) Get(key K) (value V, ok bool) {
	c.lock.Lock()
	value, ok = c.lru.Get(key)
	c.lock.Unlock()
	return value, ok
}

// Contains 检查一个键是否在缓存中，但不更新其最近使用状态
func (c *Cache[K, V]) Contains(key K) bool {
	c.lock.RLock()
	containKey := c.lru.Contains(key)
	c.lock.RUnlock()
	return containKey
}

// Peek 返回键对应的值（如果没有找到则返回未定义值），不更新其最近使用状态
func (c *Cache[K, V]) Peek(key K) (value V, ok bool) {
	c.lock.RLock()
	value, ok = c.lru.Peek(key)
	c.lock.RUnlock()
	return value, ok
}

// ContainsOrAdd 检查一个键是否在缓存中，不更新其最近使用状态，如果不存在则添加该值
// 返回是否找到以及是否发生逐出
func (c *Cache[K, V]) ContainsOrAdd(key K, value V) (ok, evicted bool) {
	var k K
	var v V
	c.lock.Lock()
	if c.lru.Contains(key) {
		c.lock.Unlock()
		return true, false
	}
	evicted = c.lru.Add(key, value)
	if c.onEvictedCB != nil && evicted {
		k, v = c.evictedKeys[0], c.evictedVals[0]
		c.evictedKeys, c.evictedVals = c.evictedKeys[:0], c.evictedVals[:0]
	}
	c.lock.Unlock()
	if c.onEvictedCB != nil && evicted {
		c.onEvictedCB(k, v)
	}
	return false, evicted
}

// PeekOrAdd 检查一个键是否在缓存中，不更新其最近使用状态，如果不存在则添加该值
// 返回之前的值，是否找到以及是否发生逐出
func (c *Cache[K, V]) PeekOrAdd(key K, value V) (previous V, ok, evicted bool) {
	var k K
	var v V
	c.lock.Lock()
	previous, ok = c.lru.Peek(key)
	if ok {
		c.lock.Unlock()
		return previous, true, false
	}
	evicted = c.lru.Add(key, value)
	if c.onEvictedCB != nil && evicted {
		k, v = c.evictedKeys[0], c.evictedVals[0]
		c.evictedKeys, c.evictedVals = c.evictedKeys[:0], c.evictedVals[:0]
	}
	c.lock.Unlock()
	if c.onEvictedCB != nil && evicted {
		c.onEvictedCB(k, v)
	}
	return
}

// Remove 从缓存中移除指定的键
func (c *Cache[K, V]) Remove(key K) (present bool) {
	var k K
	var v V
	c.lock.Lock()
	present = c.lru.Remove(key)
	if c.onEvictedCB != nil && present {
		k, v = c.evictedKeys[0], c.evictedVals[0]
		c.evictedKeys, c.evictedVals = c.evictedKeys[:0], c.evictedVals[:0]
	}
	c.lock.Unlock()
	if c.onEvictedCB != nil && present {
		c.onEvictedCB(k, v)
	}
	return
}

// Resize 调整缓存的大小
func (c *Cache[K, V]) Resize(size int) (evicted int) {
	var ks []K
	var vs []V
	c.lock.Lock()
	evicted = c.lru.Resize(size)
	if c.onEvictedCB != nil && evicted > 0 {
		ks, vs = c.evictedKeys, c.evictedVals
		c.initEvictBuffers()
	}
	c.lock.Unlock()
	if c.onEvictedCB != nil && evicted > 0 {
		for i := 0; i < len(ks); i++ {
			c.onEvictedCB(ks[i], vs[i])
		}
	}
	return evicted
}

// RemoveOldest 移除缓存中最旧的项
func (c *Cache[K, V]) RemoveOldest() (key K, value V, ok bool) {
	var k K
	var v V
	c.lock.Lock()
	key, value, ok = c.lru.RemoveOldest()
	if c.onEvictedCB != nil && ok {
		k, v = c.evictedKeys[0], c.evictedVals[0]
		c.evictedKeys, c.evictedVals = c.evictedKeys[:0], c.evictedVals[:0]
	}
	c.lock.Unlock()
	if c.onEvictedCB != nil && ok {
		c.onEvictedCB(k, v)
	}
	return
}

// GetOldest 返回缓存中最旧的键值对
func (c *Cache[K, V]) GetOldest() (key K, value V, ok bool) {
	c.lock.RLock()
	key, value, ok = c.lru.GetOldest()
	c.lock.RUnlock()
	return
}

// Keys 返回缓存中所有键的切片，从最旧到最新排序
func (c *Cache[K, V]) Keys() []K {
	c.lock.RLock()
	keys := c.lru.Keys()
	c.lock.RUnlock()
	return keys
}

// Values 返回缓存中所有值的切片，从最旧到最新排序
func (c *Cache[K, V]) Values() []V {
	c.lock.RLock()
	values := c.lru.Values()
	c.lock.RUnlock()
	return values
}

// Len 返回缓存中的项数
func (c *Cache[K, V]) Len() int {
	c.lock.RLock()
	length := c.lru.Len()
	c.lock.RUnlock()
	return length
}

// Cap 返回缓存的容量
func (c *Cache[K, V]) Cap() int {
	return c.lru.Cap()
}

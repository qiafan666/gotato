package lru

import (
	"hash/fnv"
	"sync/atomic"
	"time"
)

// Cache[T] 表示LRU缓存
type Cache[T any] struct {
	*Configuration[T]                 // 缓存配置
	control                           // 控制通道
	list              *List[*Item[T]] // 缓存列表
	size              int64           // 缓存大小
	buckets           []*bucket[T]    // 桶数组
	bucketMask        uint32          // 桶掩码
	deletables        chan *Item[T]   // 待删除通道
	promotables       chan *Item[T]   // 待提升通道
}

// New 通过指定的配置创建一个新的缓存
// 参见 lru.Configure() 以创建配置
func New[T any](config *Configuration[T]) *Cache[T] {
	c := &Cache[T]{
		list:          NewList[*Item[T]](),
		Configuration: config,
		control:       newControl(),
		bucketMask:    uint32(config.buckets) - 1,
		buckets:       make([]*bucket[T], config.buckets),
		deletables:    make(chan *Item[T], config.deleteBuffer),
		promotables:   make(chan *Item[T], config.promoteBuffer),
	}
	for i := 0; i < config.buckets; i++ {
		c.buckets[i] = &bucket[T]{
			lookup: make(map[string]*Item[T]),
		}
	}
	go c.worker() // 启动工作协程
	return c
}

// ItemCount 返回缓存中的项数
func (c *Cache[T]) ItemCount() int {
	count := 0
	for _, b := range c.buckets {
		count += b.itemCount()
	}
	return count
}

// DeletePrefix 删除指定前缀的项，并返回删除的项数
func (c *Cache[T]) DeletePrefix(prefix string) int {
	count := 0
	for _, b := range c.buckets {
		count += b.deletePrefix(prefix, c.deletables)
	}
	return count
}

// DeleteFunc 根据匹配函数删除项，并返回删除的项数
func (c *Cache[T]) DeleteFunc(matches func(key string, item *Item[T]) bool) int {
	count := 0
	for _, b := range c.buckets {
		count += b.deleteFunc(matches, c.deletables)
	}
	return count
}

// ForEachFunc 对缓存中的每一项执行指定的函数
func (c *Cache[T]) ForEachFunc(matches func(key string, item *Item[T]) bool) {
	for _, b := range c.buckets {
		if !b.forEachFunc(matches) {
			break
		}
	}
}

// Get 从缓存中获取指定键的项，如果项不存在返回nil
// 如果项已过期，也会返回，可以通过 item.Expired() 来判断是否过期
func (c *Cache[T]) Get(key string) *Item[T] {
	item := c.bucket(key).get(key)
	if item == nil {
		return nil
	}
	if !item.Expired() {
		select {
		case c.promotables <- item:
		default:
		}
	}
	return item
}

// GetWithoutPromote 从缓存中获取指定键的项，不提升该项的位置
func (c *Cache[T]) GetWithoutPromote(key string) *Item[T] {
	return c.bucket(key).get(key)
}

// TrackingGet 使用跟踪方式获取指定键的项
func (c *Cache[T]) TrackingGet(key string) TrackedItem[T] {
	item := c.Get(key)
	if item == nil {
		return nil
	}
	item.track()
	return item
}

// TrackingSet 使用跟踪方式设置项，并返回跟踪引用
func (c *Cache[T]) TrackingSet(key string, value T, duration time.Duration) TrackedItem[T] {
	return c.set(key, value, duration, true)
}

// Set 设置指定键的项及其过期时间
func (c *Cache[T]) Set(key string, value T, duration time.Duration) {
	c.set(key, value, duration, false)
}

// Setnx 仅当项不存在时设置指定键的项及其过期时间
func (c *Cache[T]) Setnx(key string, value T, duration time.Duration) {
	c.bucket(key).setnx(key, value, duration, false)
}

// Replace 如果项存在，则替换其值，否则不设置
// 返回 true 表示替换成功，false 表示项不存在
func (c *Cache[T]) Replace(key string, value T) bool {
	item := c.bucket(key).get(key)
	if item == nil {
		return false
	}
	c.Set(key, value, item.TTL())
	return true
}

// Extend 如果项存在，则延长其过期时间
// 返回 true 表示延长成功，false 表示项不存在
func (c *Cache[T]) Extend(key string, duration time.Duration) bool {
	item := c.bucket(key).get(key)
	if item == nil {
		return false
	}

	item.Extend(duration)
	return true
}

// Fetch 尝试从缓存中获取项，如果不存在或已过期则调用 fetch 方法获取值
// 如果 fetch 返回错误，则不会缓存值，并将错误返回给调用方
func (c *Cache[T]) Fetch(key string, duration time.Duration, fetch func() (T, error)) (*Item[T], error) {
	item := c.Get(key)
	if item != nil && !item.Expired() {
		return item, nil
	}
	value, err := fetch()
	if err != nil {
		return nil, err
	}
	return c.set(key, value, duration, false), nil
}

// Delete 从缓存中删除指定键的项，返回 true 表示删除成功，false 表示项不存在
func (c *Cache[T]) Delete(key string) bool {
	item := c.bucket(key).delete(key)
	if item != nil {
		c.deletables <- item
		return true
	}
	return false
}

func (c *Cache[T]) set(key string, value T, duration time.Duration, track bool) *Item[T] {
	item, existing := c.bucket(key).set(key, value, duration, track)
	if existing != nil {
		c.deletables <- existing
	}
	c.promotables <- item
	return item
}

func (c *Cache[T]) bucket(key string) *bucket[T] {
	h := fnv.New32a()
	h.Write([]byte(key))
	return c.buckets[h.Sum32()&c.bucketMask]
}

func (c *Cache[T]) halted(fn func()) {
	c.halt()
	defer c.unhalt()
	fn()
}

func (c *Cache[T]) halt() {
	for _, bucket := range c.buckets {
		bucket.Lock()
	}
}

func (c *Cache[T]) unhalt() {
	for _, bucket := range c.buckets {
		bucket.Unlock()
	}
}

func (c *Cache[T]) worker() {
	dropped := 0
	cc := c.control

	promoteItem := func(item *Item[T]) {
		if c.doPromote(item) && c.size > c.maxSize {
			dropped += c.gc()
		}
	}

	for {
		select {
		case item := <-c.promotables:
			promoteItem(item)
		case item := <-c.deletables:
			c.doDelete(item)
		case control := <-cc:
			switch msg := control.(type) {
			case controlStop:
				goto drain
			case controlGetDropped:
				msg.res <- dropped
				dropped = 0
			case controlSetMaxSize:
				c.maxSize = msg.size
				if c.size > c.maxSize {
					dropped += c.gc()
				}
				msg.done <- struct{}{}
			case controlClear:
				c.halted(func() {
					promotables := c.promotables
					for len(promotables) > 0 {
						<-promotables
					}
					deletables := c.deletables
					for len(deletables) > 0 {
						<-deletables
					}

					for _, bucket := range c.buckets {
						bucket.clear()
					}
					c.size = 0
					c.list = NewList[*Item[T]]()
				})
				msg.done <- struct{}{}
			case controlGetSize:
				msg.res <- c.size
			case controlGC:
				dropped += c.gc()
				msg.done <- struct{}{}
			case controlSyncUpdates:
				doAllPendingPromotesAndDeletes(c.promotables, promoteItem, c.deletables, c.doDelete)
				msg.done <- struct{}{}
			}
		}
	}

drain:
	for {
		select {
		case item := <-c.deletables:
			c.doDelete(item)
		default:
			return
		}
	}
}

// doAllPendingPromotesAndDeletes 用于执行所有待处理的提升和删除操作
func doAllPendingPromotesAndDeletes[T any](
	promotables <-chan *Item[T],
	promoteFn func(*Item[T]),
	deletables <-chan *Item[T],
	deleteFn func(*Item[T]),
) {
doAllPromotes:
	for {
		select {
		case item := <-promotables:
			promoteFn(item)
		default:
			break doAllPromotes
		}
	}
doAllDeletes:
	for {
		select {
		case item := <-deletables:
			deleteFn(item)
		default:
			break doAllDeletes
		}
	}
}

func (c *Cache[T]) doDelete(item *Item[T]) {
	if item.node == nil {
		item.promotions = -2
	} else {
		c.size -= item.size
		if c.onDelete != nil {
			c.onDelete(item)
		}
		c.list.Remove(item.node)
		item.node = nil
		item.promotions = -2
	}
}

func (c *Cache[T]) doPromote(item *Item[T]) bool {
	// 已删除
	if item.promotions == -2 {
		return false
	}
	if item.node != nil { // 不是新项
		if item.shouldPromote(c.getsPerPromote) {
			c.list.MoveToFront(item.node)
			item.promotions = 0
		}
		return false
	}

	c.size += item.size
	item.node = c.list.Insert(item)
	return true
}

func (c *Cache[T]) gc() int {
	dropped := 0
	node := c.list.Tail

	itemsToPrune := int64(c.itemsToPrune)
	if min := c.size - c.maxSize; min > itemsToPrune {
		itemsToPrune = min
	}

	for i := int64(0); i < itemsToPrune; i++ {
		if node == nil {
			return dropped
		}
		prev := node.Prev
		item := node.Value
		if !c.tracking || atomic.LoadInt32(&item.refCount) == 0 {
			c.bucket(item.key).delete(item.key)
			c.size -= item.size
			c.list.Remove(node)
			if c.onDelete != nil {
				c.onDelete(item)
			}
			dropped += 1
			item.node = nil
			item.promotions = -2
		}
		node = prev
	}
	return dropped
}

package lru

import (
	"hash/fnv"
	"sync/atomic"
	"time"
)

// LayeredCache 是一个分层缓存
type LayeredCache[T any] struct {
	*Configuration[T]
	control                         // 控制通道
	list        *List[*Item[T]]     // 缓存项链表
	buckets     []*layeredBucket[T] // 分桶
	bucketMask  uint32              // 分桶掩码
	size        int64               // 缓存大小
	deletables  chan *Item[T]       // 待删除缓存项通道
	promotables chan *Item[T]       // 待提升缓存项通道
}

// Layered 创建一个新的分层缓存
func Layered[T any](config *Configuration[T]) *LayeredCache[T] {
	c := &LayeredCache[T]{
		list:          NewList[*Item[T]](), // 初始化缓存项链表
		Configuration: config,
		control:       newControl(), // 初始化控制通道
		bucketMask:    uint32(config.buckets) - 1,
		buckets:       make([]*layeredBucket[T], config.buckets),
		deletables:    make(chan *Item[T], config.deleteBuffer),  // 初始化待删除缓存项通道
		promotables:   make(chan *Item[T], config.promoteBuffer), // 初始化待提升缓存项通道
	}
	// 初始化分桶
	for i := 0; i < config.buckets; i++ {
		c.buckets[i] = &layeredBucket[T]{
			buckets: make(map[string]*bucket[T]),
		}
	}
	// 启动工作协程
	go c.worker()
	return c
}

// ItemCount 返回缓存项的数量
func (c *LayeredCache[T]) ItemCount() int {
	count := 0
	for _, b := range c.buckets {
		count += b.itemCount()
	}
	return count
}

// Get 获取缓存项
func (c *LayeredCache[T]) Get(primary, secondary string) *Item[T] {
	item := c.bucket(primary).get(primary, secondary)
	if item == nil {
		return nil
	}
	if item.expires > time.Now().UnixNano() {
		select {
		case c.promotables <- item:
		default:
		}
	}
	return item
}

// GetWithoutPromote 获取缓存项（不触发提升操作）
func (c *LayeredCache[T]) GetWithoutPromote(primary, secondary string) *Item[T] {
	return c.bucket(primary).get(primary, secondary)
}

// ForEachFunc 遍历缓存项并执行函数
func (c *LayeredCache[T]) ForEachFunc(primary string, matches func(key string, item *Item[T]) bool) {
	c.bucket(primary).forEachFunc(primary, matches)
}

// GetOrCreateSecondaryCache 获取或创建二级缓存
func (c *LayeredCache[T]) GetOrCreateSecondaryCache(primary string) *SecondaryCache[T] {
	primaryBkt := c.bucket(primary)
	bkt := primaryBkt.getSecondaryBucket(primary)
	primaryBkt.Lock()
	if bkt == nil {
		bkt = &bucket[T]{lookup: make(map[string]*Item[T])}
		primaryBkt.buckets[primary] = bkt
	}
	primaryBkt.Unlock()
	return &SecondaryCache[T]{
		bucket: bkt,
		pCache: c,
	}
}

// TrackingGet 获取并追踪缓存项
func (c *LayeredCache[T]) TrackingGet(primary, secondary string) TrackedItem[T] {
	item := c.Get(primary, secondary)
	if item == nil {
		return nil
	}
	item.track()
	return item
}

// TrackingSet 设置缓存项并追踪
func (c *LayeredCache[T]) TrackingSet(primary, secondary string, value T, duration time.Duration) TrackedItem[T] {
	return c.set(primary, secondary, value, duration, true)
}

// Set 设置缓存项
func (c *LayeredCache[T]) Set(primary, secondary string, value T, duration time.Duration) {
	c.set(primary, secondary, value, duration, false)
}

// Replace 替换缓存项
func (c *LayeredCache[T]) Replace(primary, secondary string, value T) bool {
	item := c.bucket(primary).get(primary, secondary)
	if item == nil {
		return false
	}
	c.Set(primary, secondary, value, item.TTL())
	return true
}

// Fetch 获取或者创建缓存项（如果不存在）
func (c *LayeredCache[T]) Fetch(primary, secondary string, duration time.Duration, fetch func() (T, error)) (*Item[T], error) {
	item := c.Get(primary, secondary)
	if item != nil {
		return item, nil
	}
	value, err := fetch()
	if err != nil {
		return nil, err
	}
	return c.set(primary, secondary, value, duration, false), nil
}

// Delete 删除缓存项
func (c *LayeredCache[T]) Delete(primary, secondary string) bool {
	item := c.bucket(primary).delete(primary, secondary)
	if item != nil {
		c.deletables <- item
		return true
	}
	return false
}

// DeleteAll 删除指定主键下的所有缓存项
func (c *LayeredCache[T]) DeleteAll(primary string) bool {
	return c.bucket(primary).deleteAll(primary, c.deletables)
}

// DeletePrefix 删除指定主键下以指定前缀开头的缓存项
func (c *LayeredCache[T]) DeletePrefix(primary, prefix string) int {
	return c.bucket(primary).deletePrefix(primary, prefix, c.deletables)
}

// DeleteFunc 根据指定条件删除缓存项
func (c *LayeredCache[T]) DeleteFunc(primary string, matches func(key string, item *Item[T]) bool) int {
	return c.bucket(primary).deleteFunc(primary, matches, c.deletables)
}

// set 设置缓存项
func (c *LayeredCache[T]) set(primary, secondary string, value T, duration time.Duration, track bool) *Item[T] {
	item, existing := c.bucket(primary).set(primary, secondary, value, duration, track)
	if existing != nil {
		c.deletables <- existing
	}
	c.promote(item)
	return item
}

// bucket 返回缓存项对应的分桶
func (c *LayeredCache[T]) bucket(key string) *layeredBucket[T] {
	h := fnv.New32a()
	h.Write([]byte(key))
	return c.buckets[h.Sum32()&c.bucketMask]
}

// halted 暂停缓存操作
func (c *LayeredCache[T]) halted(fn func()) {
	c.halt()
	defer c.unhalt()
	fn()
}

func (c *LayeredCache[T]) halt() {
	for _, bucket := range c.buckets {
		bucket.Lock()
	}
}

func (c *LayeredCache[T]) unhalt() {
	for _, bucket := range c.buckets {
		bucket.Unlock()
	}
}

func (c *LayeredCache[T]) promote(item *Item[T]) {
	c.promotables <- item
}

func (c *LayeredCache[T]) worker() {
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
				promotables := c.promotables
				for len(promotables) > 0 {
					<-promotables
				}
				deletables := c.deletables
				for len(deletables) > 0 {
					<-deletables
				}

				c.halted(func() {
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

func (c *LayeredCache[T]) doDelete(item *Item[T]) {
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

func (c *LayeredCache[T]) doPromote(item *Item[T]) bool {
	// deleted before it ever got promoted
	if item.promotions == -2 {
		return false
	}
	if item.node != nil { //not a new item
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

func (c *LayeredCache[T]) gc() int {
	node := c.list.Tail
	dropped := 0
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
			c.bucket(item.group).delete(item.group, item.key)
			c.size -= item.size
			c.list.Remove(node)
			if c.onDelete != nil {
				c.onDelete(item)
			}
			item.node = nil
			item.promotions = -2
			dropped += 1
		}
		node = prev
	}
	return dropped
}

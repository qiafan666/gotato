package glru

import (
	"fmt"
	"sync/atomic"
	"time"
)

// Sized 接口定义了一个获取大小的方法
type Sized interface {
	Size() int64
}

// TrackedItem 接口定义了一个被追踪的缓存项应该实现的方法集合
type TrackedItem[T any] interface {
	Value() T                      // 返回缓存项的值
	Release()                      // 释放缓存项，减少引用计数
	Expired() bool                 // 检查缓存项是否过期
	TTL() time.Duration            // 返回缓存项的剩余时间
	Expires() time.Time            // 返回缓存项的过期时间
	Extend(duration time.Duration) // 延长缓存项的过期时间
}

// Item 是LRU缓存中的一个项目
type Item[T any] struct {
	key        string          // 键
	group      string          // 分组
	promotions int32           // 提升次数
	refCount   int32           // 引用计数
	expires    int64           // 过期时间（纳秒）
	size       int64           // 大小
	value      T               // 值
	node       *Node[*Item[T]] // 链表节点
}

// newItem 创建一个新的缓存项
func newItem[T any](key string, value T, expires int64, track bool) *Item[T] {
	size := int64(1)

	// 如果值实现了 Sized 接口，则获取其大小
	if sized, ok := (interface{})(value).(Sized); ok {
		size = sized.Size()
	}
	item := &Item[T]{
		key:        key,
		value:      value,
		promotions: 0,
		size:       size,
		expires:    expires,
	}
	// 如果需要追踪，则设置初始引用计数为1
	if track {
		item.refCount = 1
	}
	return item
}

// shouldPromote 检查缓存项是否应该提升
func (i *Item[T]) shouldPromote(getsPerPromote int32) bool {
	i.promotions += 1
	return i.promotions == getsPerPromote
}

// Key 返回缓存项的键
func (i *Item[T]) Key() string {
	return i.key
}

// Value 返回缓存项的值
func (i *Item[T]) Value() T {
	return i.value
}

// track 增加缓存项的引用计数
func (i *Item[T]) track() {
	atomic.AddInt32(&i.refCount, 1)
}

// Release 释放缓存项，减少引用计数
func (i *Item[T]) Release() {
	atomic.AddInt32(&i.refCount, -1)
}

// Expired 检查缓存项是否过期
func (i *Item[T]) Expired() bool {
	expires := atomic.LoadInt64(&i.expires)
	return expires < time.Now().UnixNano()
}

// TTL 返回缓存项的剩余时间
func (i *Item[T]) TTL() time.Duration {
	expires := atomic.LoadInt64(&i.expires)
	return time.Nanosecond * time.Duration(expires-time.Now().UnixNano())
}

// Expires 返回缓存项的过期时间
func (i *Item[T]) Expires() time.Time {
	expires := atomic.LoadInt64(&i.expires)
	return time.Unix(0, expires)
}

// Extend 延长缓存项的过期时间
func (i *Item[T]) Extend(duration time.Duration) {
	atomic.StoreInt64(&i.expires, time.Now().Add(duration).UnixNano())
}

// String 返回缓存项的字符串表示形式
func (i *Item[T]) String() string {
	return fmt.Sprintf("Item(%v)", i.value)
}

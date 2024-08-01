package glru

import (
	"strings"
	"sync"
	"time"
)

// bucket[T] 表示LRU缓存中的一个桶，存储键值对
type bucket[T any] struct {
	sync.RWMutex                     // 读写锁，用于并发安全
	lookup       map[string]*Item[T] // 存储键值对的映射表
}

// itemCount 返回桶中存储的键值对数量
func (b *bucket[T]) itemCount() int {
	b.RLock()
	defer b.RUnlock()
	return len(b.lookup)
}

// forEachFunc 遍历桶中的键值对，执行指定的函数，如果函数返回 false 则停止遍历
func (b *bucket[T]) forEachFunc(matches func(key string, item *Item[T]) bool) bool {
	lookup := b.lookup
	b.RLock()
	defer b.RUnlock()
	for key, item := range lookup {
		if !matches(key, item) {
			return false
		}
	}
	return true
}

// get 根据键获取桶中的值
func (b *bucket[T]) get(key string) *Item[T] {
	b.RLock()
	defer b.RUnlock()
	return b.lookup[key]
}

// setnx 根据键设置值，如果键已存在则返回已存在的值，否则设置新值并返回
func (b *bucket[T]) setnx(key string, value T, duration time.Duration, track bool) *Item[T] {
	b.RLock()
	item := b.lookup[key]
	b.RUnlock()
	if item != nil {
		return item
	}

	expires := time.Now().Add(duration).UnixNano()
	newItemInfo := newItem(key, value, expires, track)

	b.Lock()
	defer b.Unlock()

	// 再次检查，确保并发安全
	item = b.lookup[key]
	if item != nil {
		return item
	}

	b.lookup[key] = newItemInfo
	return newItemInfo
}

// set 根据键设置值，返回设置后的新值和可能已存在的旧值
func (b *bucket[T]) set(key string, value T, duration time.Duration, track bool) (*Item[T], *Item[T]) {
	expires := time.Now().Add(duration).UnixNano()
	item := newItem(key, value, expires, track)
	b.Lock()
	existing := b.lookup[key]
	b.lookup[key] = item
	b.Unlock()
	return item, existing
}

// delete 根据键删除桶中的值，并返回被删除的值
func (b *bucket[T]) delete(key string) *Item[T] {
	b.Lock()
	item := b.lookup[key]
	delete(b.lookup, key)
	b.Unlock()
	return item
}

// deleteFunc 根据匹配函数删除桶中的值，并返回删除的值的数量
func (b *bucket[T]) deleteFunc(matches func(key string, item *Item[T]) bool, deletables chan *Item[T]) int {
	lookup := b.lookup
	items := make([]*Item[T], 0)

	b.RLock()
	for key, item := range lookup {
		if matches(key, item) {
			deletables <- item
			items = append(items, item)
		}
	}
	b.RUnlock()

	if len(items) == 0 {
		// 避免加写锁
		return 0
	}

	b.Lock()
	for _, item := range items {
		delete(lookup, item.key)
	}
	b.Unlock()
	return len(items)
}

// deletePrefix 根据前缀删除桶中的值，并返回删除的值的数量
func (b *bucket[T]) deletePrefix(prefix string, deletables chan *Item[T]) int {
	return b.deleteFunc(func(key string, item *Item[T]) bool {
		return strings.HasPrefix(key, prefix)
	}, deletables)
}

// clear 清空桶中的所有键值对
// 预期调用者已经获得了写锁
func (b *bucket[T]) clear() {
	for _, item := range b.lookup {
		item.promotions = -2
	}
	b.lookup = make(map[string]*Item[T])
}

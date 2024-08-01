package gcache

import (
	"container/list"
	"sync"
)

// LRUCache Least Recently Used，最近最少使用
type LRUCache struct {
	items     map[string]*list.Element // 用于快速检索缓存中的条目，键是字符串，值是双向链表的元素指针
	evictList *list.List               // 双向链表，按照最近使用顺序存储缓存中的条目
	capacity  int                      // 缓存的最大容量
	mu        sync.Mutex               // 独占锁，用于保护并发访问缓存的安全
}

// entry 是LRUCache中每个条目的结构体，包括键和值
type entry struct {
	key   string      // 键
	value interface{} // 值的类型是任意类型，使用空接口(interface{})表示
}

// NewLRUCache 创建LRUCache对象，并指定缓存的容量
func NewLRUCache(capacity int) LRUCache {
	return LRUCache{
		items:     make(map[string]*list.Element, 2), // 初始化items为map类型，键为字符串，值为双向链表元素的指针
		evictList: list.New(),                        // 初始化evictList为双向链表
		capacity:  capacity,                          // 设置缓存的容量
	}
}

// Get 根据键从LRU缓存中获取值
func (cache *LRUCache) Get(key string) interface{} {
	cache.mu.Lock()         // 加读锁，允许多个读取者并发访问
	defer cache.mu.Unlock() // 函数执行完毕后释放读锁

	// 检查键是否存在于缓存中
	ent, ok := cache.items[key]
	if ok {
		cache.evictList.MoveToFront(ent) // 将对应条目移到链表头部，表示最近使用
		return ent.Value.(*entry).value  // 返回条目的值
	}

	return nil // 如果键不存在于缓存中，则返回nil
}

// GetFrontEntries 获取缓存中前n个键值对，如果num为-1，则获取所有键值对
func (cache *LRUCache) GetFrontEntries(num int) map[string]interface{} {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	result := make(map[string]interface{})
	if num == -1 {
		// 遍历双向链表，获取所有键值对
		for ent := cache.evictList.Front(); ent != nil; ent = ent.Next() {
			key := ent.Value.(*entry).key
			value := ent.Value.(*entry).value
			result[key] = value
		}
	} else {
		// 遍历双向链表，获取前n个键值对
		count := 0
		for ent := cache.evictList.Front(); ent != nil && count < num; ent = ent.Next() {
			key := ent.Value.(*entry).key
			value := ent.Value.(*entry).value
			result[key] = value
			count++
		}
	}
	return result
}

// GetBackEntries 获取缓存中后n个键值对，如果n为-1，则获取所有键值对
func (cache *LRUCache) GetBackEntries(n int) map[string]interface{} {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	result := make(map[string]interface{})
	if n == -1 {
		// 遍历双向链表，获取所有键值对
		for ent := cache.evictList.Back(); ent != nil; ent = ent.Prev() {
			key := ent.Value.(*entry).key
			value := ent.Value.(*entry).value
			result[key] = value
		}
	} else {
		// 遍历双向链表，获取后n个键值对
		count := 0
		for ent := cache.evictList.Back(); ent != nil && count < n; ent = ent.Prev() {
			key := ent.Value.(*entry).key
			value := ent.Value.(*entry).value
			result[key] = value
			count++
		}
	}
	return result
}

// Put 将键值对添加到LRU缓存中
func (cache *LRUCache) Put(key string, value interface{}) {
	cache.mu.Lock()         // 加锁，保护并发访问
	defer cache.mu.Unlock() // 函数执行完毕后释放锁

	// 检查键是否已经存在于缓存中
	if ent, ok := cache.items[key]; ok {
		cache.evictList.MoveToFront(ent) // 将对应条目移到链表头部，表示最近使用
		ent.Value.(*entry).value = value // 更新条目的值
		return
	}

	// 如果缓存容量已达上限，则移除最久未使用的条目
	if cache.evictList.Len() == cache.capacity {
		ent := cache.evictList.Back() // 获取最久未使用的条目
		kv := ent.Value.(*entry)      // 获取条目的键值对
		cache.evictList.Remove(ent)   // 从链表中移除最久未使用的条目
		delete(cache.items, kv.key)   // 从items中删除最久未使用的条目
	}

	// 创建新的条目，并添加到缓存中
	entryValue := cache.evictList.PushFront(&entry{key, value}) // 在链表头部添加新条目
	cache.items[key] = entryValue                               // 在items中添加新条目
}

// RemoveKey 从LRU缓存中删除指定的键值对
func (cache *LRUCache) RemoveKey(key string) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if ent, ok := cache.items[key]; ok {
		delete(cache.items, key)
		cache.evictList.Remove(ent)
	}
}

// RemoveBackEntries 删除LRU缓存中的后面几个条目
func (cache *LRUCache) RemoveBackEntries(n int) {
	cache.mu.Lock()         // 加锁，保护并发访问
	defer cache.mu.Unlock() // 函数执行完毕后释放锁

	// 删除后面n个条目
	for i := 0; i < n && cache.evictList.Len() > 0; i++ {
		ent := cache.evictList.Back() // 获取最久未使用的条目
		key := ent.Value.(*entry).key // 获取条目的键
		cache.evictList.Remove(ent)   // 从链表中移除最久未使用的条目
		delete(cache.items, key)      // 从items中删除最久未使用的条目
	}
}

// RemoveFrontEntries 根据数量删除LRU缓存中的头部条目
func (cache *LRUCache) RemoveFrontEntries(n int) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	// 如果要删除的数量大于等于LRU缓存的当前容量，则直接清空LRU缓存
	if n >= cache.evictList.Len() {
		cache.Clear()
		return
	}

	// 删除LRU缓存中的头部条目，直到达到指定的数量
	for i := 0; i < n; i++ {
		ent := cache.evictList.Front()
		if ent != nil {
			key := ent.Value.(*entry).key
			delete(cache.items, key)
			cache.evictList.Remove(ent)
		}
	}
}

// Size 返回LRU缓存中的键值对数量
func (cache *LRUCache) Size() int {
	cache.mu.Lock()         // 加读锁，允许多个读取者并发访问
	defer cache.mu.Unlock() // 函数执行完毕后释放读锁

	return len(cache.items)
}

// Clear 清空LRU缓存中的所有键值对
func (cache *LRUCache) Clear() {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.items = make(map[string]*list.Element)
	cache.evictList.Init()
}

// GetCapacity 获取LRU缓存的最大容量
func (cache *LRUCache) GetCapacity() int {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	return cache.capacity
}

// SetCapacity 设置LRU缓存的最大容量
func (cache *LRUCache) SetCapacity(capacity int) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.capacity = capacity
}

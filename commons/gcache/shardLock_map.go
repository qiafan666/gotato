package gcache

import (
	"encoding/json"
	"fmt"
	"github.com/qiafan666/gotato/commons/gson"
	"sync"
)

// concurrent_map可当做缓存数据库
var defaultShardCount = 16 // 默认分片数

// Stringer 接口定义，需要同时具备fmt.Stringer和可比较性
type Stringer interface {
	fmt.Stringer
	comparable
}

// ShardLockMap 结构定义，使用泛型
// K 和 V 分别代表键和值的类型
type ShardLockMap[K comparable, V any] struct {
	shardCount int                        // 分片数
	shards     []*shardLockMapShard[K, V] // 分片数组
	sharding   func(key K) uint32         // 分片函数
}

// shardLockMapShard 是分片的具体实现
type shardLockMapShard[K comparable, V any] struct {
	items        map[K]V
	sync.RWMutex // 读写锁
}

// NewShardLockMap 创建一个以string为键的并发Map
func NewShardLockMap[V any](shardCounts ...int) *ShardLockMap[string, V] {
	if len(shardCounts) > 0 {
		return create[string, V](fnv32, shardCounts[0])
	}
	return create[string, V](fnv32, defaultShardCount)
}

// NewStringer 创建一个以Stringer接口实现者为键的并发Map
func NewStringer[K Stringer, V any](shardCounts ...int) *ShardLockMap[K, V] {
	if len(shardCounts) > 0 {
		return create[K, V](strFnv32[K], shardCounts[0])
	}
	return create[K, V](strFnv32[K], defaultShardCount)
}

// NewWithCustomShardingFunction 允许使用自定义分片函数
func NewWithCustomShardingFunction[K comparable, V any](sharding func(key K) uint32, shardCounts ...int) *ShardLockMap[K, V] {
	if len(shardCounts) > 0 {
		return create[K, V](sharding, shardCounts[0])
	}
	return create[K, V](sharding, defaultShardCount)
}

// 创建一个新的 ShardLockMap，限定于包内
func create[K comparable, V any](sharding func(key K) uint32, defaultShardCount int) *ShardLockMap[K, V] {
	newMap := &ShardLockMap[K, V]{
		shardCount: defaultShardCount,
		sharding:   sharding,
		shards:     make([]*shardLockMapShard[K, V], defaultShardCount),
	}
	for i := 0; i < defaultShardCount; i++ {
		newMap.shards[i] = &shardLockMapShard[K, V]{items: make(map[K]V)}
	}
	return newMap
}

// 根据键获取对应的分片
func (m *ShardLockMap[K, V]) getShard(key K) *shardLockMapShard[K, V] {
	return m.shards[uint(m.sharding(key))%uint(m.shardCount)]
}

// MSet 批量设置键值对
func (m *ShardLockMap[K, V]) MSet(data map[K]V) {
	for key, value := range data {
		shard := m.getShard(key)
		shard.Lock()
		shard.items[key] = value
		shard.Unlock()
	}
}

// Set 设置单个键值对
func (m *ShardLockMap[K, V]) Set(key K, value V) {
	shard := m.getShard(key)
	shard.Lock()
	shard.items[key] = value
	shard.Unlock()
}

type upsertCb[V any] func(exist bool, valueInMap V, newValue V) V

// Upsert 插入或更新一个元素
func (m *ShardLockMap[K, V]) Upsert(key K, value V, cb upsertCb[V]) (res V) {
	shard := m.getShard(key)
	shard.Lock()
	v, ok := shard.items[key]
	res = cb(ok, v, value)
	shard.items[key] = res
	shard.Unlock()
	return res
}

// SetIfAbsent 检查键是否存在，如果不存在，则设置新值
// 如果键已存在，返回键对应的值和true；如果键不存在，设置新值并返回nil和false。
func (m *ShardLockMap[K, V]) SetIfAbsent(key K, value V) (V, bool) {
	shard := m.getShard(key)
	shard.Lock()
	defer shard.Unlock()

	existingValue, exists := shard.items[key]
	if exists {
		return existingValue, true
	}

	shard.items[key] = value
	return value, false
}

// Get 检索指定键的值
func (m *ShardLockMap[K, V]) Get(key K) (V, bool) {
	shard := m.getShard(key)
	shard.RLock()
	val, ok := shard.items[key]
	shard.RUnlock()
	return val, ok
}

// Count 返回Map中的元素总数
func (m *ShardLockMap[K, V]) Count() int {
	count := 0
	for _, shard := range m.shards {
		shard.RLock()
		count += len(shard.items)
		shard.RUnlock()
	}
	return count
}

// Has 检查键是否存在于Map中
func (m *ShardLockMap[K, V]) Has(key K) bool {
	shard := m.getShard(key)
	shard.RLock()
	_, ok := shard.items[key]
	shard.RUnlock()
	return ok
}

// Remove 移除指定键
func (m *ShardLockMap[K, V]) Remove(key K) {
	shard := m.getShard(key)
	shard.Lock()
	delete(shard.items, key)
	shard.Unlock()
}

// removeCb 定义了移除元素时的回调类型
type removeCb[K comparable, V any] func(key K, v V, exists bool) bool

// RemoveCb 执行带回调的移除操作
func (m *ShardLockMap[K, V]) RemoveCb(key K, cb removeCb[K, V]) bool {
	shard := m.getShard(key)
	shard.Lock()
	v, exists := shard.items[key]
	shouldRemove := cb(key, v, exists)
	if shouldRemove && exists {
		delete(shard.items, key)
	}
	shard.Unlock()
	return shouldRemove
}

// Pop 移除并返回指定键的值
func (m *ShardLockMap[K, V]) Pop(key K) (v V, exists bool) {
	shard := m.getShard(key)
	shard.Lock()
	v, exists = shard.items[key]
	if exists {
		delete(shard.items, key)
	}
	shard.Unlock()
	return v, exists
}

// Clear 清空Map
func (m *ShardLockMap[K, V]) Clear() {
	for _, shard := range m.shards {
		shard.Lock()
		shard.items = make(map[K]V)
		shard.Unlock()
	}
}

// Items 返回所有键值对
func (m *ShardLockMap[K, V]) Items() map[K]V {
	tmp := make(map[K]V)
	for _, shard := range m.shards {
		shard.RLock()
		for key, value := range shard.items {
			tmp[key] = value
		}
		shard.RUnlock()
	}
	return tmp
}

// MarshalJSON 序列化当前并发Map为JSON格式
func (m *ShardLockMap[K, V]) MarshalJSON() ([]byte, error) {
	items := m.Items()
	return gson.Marshal(items)
}

func strFnv32[K fmt.Stringer](key K) uint32 {
	return fnv32(key.String())
}

func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	keyLength := len(key)
	for i := 0; i < keyLength; i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}

// UnMarshalJSON 反序列化JSON格式的并发Map
// 注意：并发Map的键类型必须可以被json.Unmarshal解析，否则会导致panic
func (m *ShardLockMap[K, V]) UnMarshalJSON(b []byte) (err error) {
	tmp := make(map[K]V)

	// Unmarshal into a single map.
	if err = json.Unmarshal(b, &tmp); err != nil {
		return err
	}

	for key, val := range tmp {
		m.Set(key, val)
	}
	return nil
}

// Keys 返回所有键
func (m *ShardLockMap[K, V]) Keys() []K {
	count := m.Count()
	ch := make(chan K, count)
	go func() {
		wg := sync.WaitGroup{}
		wg.Add(m.shardCount)
		for _, shard := range m.shards {
			go func(shard *shardLockMapShard[K, V]) {
				shard.RLock()
				for key := range shard.items {
					ch <- key
				}
				shard.RUnlock()
				wg.Done()
			}(shard)
		}
		wg.Wait()
		close(ch)
	}()

	keys := make([]K, 0, count)
	for k := range ch {
		keys = append(keys, k)
	}
	return keys
}

// IterCb 是一个以键和值作为参数的回调函数类型
type IterCb func(key string, value interface{})

// IterCb 方法用于迭代并调用提供的回调函数
func (m *ShardLockMap[K, V]) IterCb(fn IterCb) {
	for idx := range m.shards {
		shard := m.shards[idx]
		shard.RLock()
		for key, value := range shard.items {
			keyStr := fmt.Sprintf("%v", key) // 转换为字符串
			fn(keyStr, value)                // 直接传递 key 和 value
		}
		shard.RUnlock()
	}
}

// KVPairs 用于存储键值对
type KVPairs[K comparable, V any] struct {
	Key   K
	Value V
}

// IterBuffered 返回一个缓冲的迭代器，可以在 for range 循环中使用。
func (m *ShardLockMap[K, V]) IterBuffered() <-chan KVPairs[K, V] {
	chanList := m.snapshot() // 使用)m.snapshot()获取通道列表
	total := 0
	for _, c := range chanList {
		total += cap(c)
	}
	ch := make(chan KVPairs[K, V], total)
	go fanIn(chanList, ch)
	return ch
}

// snapshot 返回一个包含每个分片元素的通道数组，并估计每个缓冲通道的大小。
func (m *ShardLockMap[K, V]) snapshot() []chan KVPairs[K, V] {
	chanList := make([]chan KVPairs[K, V], m.shardCount)
	wg := sync.WaitGroup{}
	wg.Add(m.shardCount)
	for index, shard := range m.shards {
		go func(index int, shard *shardLockMapShard[K, V]) {
			defer wg.Done() // 完成任务后递减计数
			shard.RLock()
			defer shard.RUnlock()
			chanList[index] = make(chan KVPairs[K, V], len(shard.items))
			// 将所有键值对发送到通道
			for key, val := range shard.items {
				chanList[index] <- KVPairs[K, V]{Key: key, Value: val}
			}
			close(chanList[index]) // 关闭通道
		}(index, shard)
	}
	wg.Wait()
	return chanList
}

// fanIn 从通道 chanList 中读取元素并写入通道 out，确保能够识别 K 和 V 的类型
func fanIn[K comparable, V any](chanList []chan KVPairs[K, V], out chan KVPairs[K, V]) {
	wg := sync.WaitGroup{}
	wg.Add(len(chanList))
	for _, ch := range chanList {
		go func(ch chan KVPairs[K, V]) {
			for t := range ch {
				out <- t
			}
			wg.Done()
		}(ch)
	}
	wg.Wait()
	close(out)
}

package gcache

import (
	"github.com/qiafan666/gotato/commons/gcommon"
	"sort"
	"sync"
)

// lfuItem 是LFU缓存中的一个项
type lfuItem struct {
	key   string      // 键
	value interface{} // 值，空接口类型可以存储任意类型的值
	freq  int         // 访问频率
}

// LFUCache least frequently used cache 最不经常使用缓存
type LFUCache struct {
	capacity int                            // 缓存容量
	items    map[string]lfuItem             // 缓存所有项
	freqMap  map[int]map[string]interface{} //跟频率挂钩的items
	ascFreq  []int
	descFreq []int
	mu       sync.Mutex // 互斥锁，确保并发安全
}

// NewLFUCache 创建一个LFU缓存
func NewLFUCache(capacity int) *LFUCache {
	return &LFUCache{
		capacity: capacity,
		items:    make(map[string]lfuItem),
		freqMap:  make(map[int]map[string]interface{}),
		ascFreq:  []int{},
		descFreq: []int{},
	}
}

// Put 将键值对放入缓存中
func (lfu *LFUCache) Put(key string, value interface{}) {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	var newLfuItem lfuItem
	if item, ok := lfu.items[key]; ok {
		item.freq++
		newLfuItem = item
	} else {
		newLfuItem = lfuItem{
			key:   key,
			value: value,
			freq:  1,
		}
	}

	//添加item
	lfu.addItem(newLfuItem)

	// 更新频率映射
	lfu.updateSort()
}

// updateSort 更新频率层级
func (lfu *LFUCache) updateSort() {

	var newDescFreq []int
	var newAscFreq []int
	for freq, _ := range lfu.freqMap {
		if len(lfu.freqMap) == 1 {
			newDescFreq = []int{freq}
			newAscFreq = []int{freq}
			break
		} else {
			newDescFreq = append(newDescFreq, freq)
			newAscFreq = append(newAscFreq, freq)
		}
	}

	lfu.descFreq = newDescFreq
	lfu.ascFreq = newAscFreq

	sort.Slice(lfu.descFreq, func(i, j int) bool {
		return lfu.descFreq[i] > lfu.descFreq[j]
	})

	// 创建一个新的 ascFreq 数组
	ascFreq := make([]int, len(lfu.descFreq))
	for i := 0; i < len(lfu.descFreq); i++ {
		ascFreq[i] = lfu.descFreq[len(lfu.descFreq)-1-i]
	}

	// 将新的 ascFreq 数组赋值给 lfu.ascFreq
	lfu.ascFreq = ascFreq
}

func (lfu *LFUCache) addItem(item lfuItem) {

	if lfu.capacity == 0 {
		return
	}

	if len(lfu.items) >= lfu.capacity {
		//删除最末尾的一项
		for s, _ := range lfu.freqMap[lfu.ascFreq[0]] {
			delete(lfu.freqMap[lfu.ascFreq[0]], s)
			delete(lfu.items, s)
			break
		}
		//如果当前最小频率的值为空则清除
		if len(lfu.freqMap[lfu.ascFreq[0]]) == 0 {
			delete(lfu.freqMap, item.freq)
			delete(lfu.items, item.key)
		}
	}

	if value, ok := lfu.items[item.key]; ok {
		delete(lfu.freqMap[value.freq], value.key)
		if len(lfu.freqMap[value.freq]) == 0 {
			delete(lfu.freqMap, value.freq)
			delete(lfu.items, value.key)
			gcommon.SliceDelete(lfu.ascFreq, value.freq)
			gcommon.SliceDelete(lfu.descFreq, value.freq)
		}
	}

	//添加新项
	if existMap, exists := lfu.freqMap[item.freq]; !exists {
		m2 := make(map[string]interface{})
		m2[item.key] = item.value
		lfu.freqMap[item.freq] = m2
	} else {
		existMap[item.key] = item.value
		lfu.freqMap[item.freq] = existMap
	}
	lfu.items[item.key] = item
}

// ContainsFreq 检查频率列表中是否包含指定频率
func (lfu *LFUCache) ContainsFreq(freq int) bool {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()
	for f, _ := range lfu.freqMap {
		if f == freq {
			return true
		}
	}
	return false
}

// Remove 移除最不经常使用的项
func (lfu *LFUCache) Remove(keys ...string) {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	for _, key := range keys {
		if item, ok := lfu.items[key]; ok {
			delete(lfu.freqMap[item.freq], item.key)

			if len(lfu.freqMap[item.freq]) == 0 {
				delete(lfu.freqMap, item.freq)
			}
		}
		delete(lfu.items, key)
	}
	lfu.updateSort()
}

// GetFrontFreq 获取前n个频率
func (lfu *LFUCache) GetFrontFreq(n int) []int {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	if n > len(lfu.descFreq) || n < 0 {
		return lfu.descFreq
	}

	return lfu.descFreq[0:n]
}

// GetBackFreq 获取后n个频率
func (lfu *LFUCache) GetBackFreq(n int) []int {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	if n > len(lfu.ascFreq) || n < 0 {
		return lfu.ascFreq
	}

	return lfu.ascFreq[0:n]
}

// KeysAndValues 获取缓存中所有键值对
func (lfu *LFUCache) KeysAndValues() map[string]interface{} {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	result := make(map[string]interface{})
	for _, item := range lfu.items {
		result[item.key] = item.value
	}
	return result
}

// Clear 清空缓存
func (lfu *LFUCache) Clear() {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	lfu.items = make(map[string]lfuItem)
	lfu.freqMap = make(map[int]map[string]interface{})
	lfu.ascFreq = []int{}
	lfu.descFreq = []int{}
}

// Contains 检查键是否存在于缓存中
func (lfu *LFUCache) Contains(key string) bool {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	_, ok := lfu.items[key]
	return ok
}

// Size 获取缓存的大小
func (lfu *LFUCache) Size() int {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	return len(lfu.items)
}

// FreqItems 返回指定频率的项
func (lfu *LFUCache) FreqItems(freq int) map[string]interface{} {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	if f, ok := lfu.freqMap[freq]; ok {
		items := make(map[string]interface{})
		for key, value := range f {
			items[key] = value
		}
		return items
	}
	return nil
}

// GetFrontFreqKVS 返回前面频率层级的所有KV，-1返回所有
func (lfu *LFUCache) GetFrontFreqKVS(num int) map[int]map[string]interface{} {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	results := make(map[int]map[string]interface{})

	count := 0
	for _, freq := range lfu.descFreq {
		if kvs, ok := lfu.freqMap[freq]; ok {
			if count >= num && num >= 0 {
				return results
			}
			results[freq] = kvs
			count++
		}
	}

	return results
}

// GetFrontEntries 返回前面数量的KV，-1返回所有
func (lfu *LFUCache) GetFrontEntries(num int) map[int]map[string]interface{} {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	results := make(map[int]map[string]interface{})

	count := 0
	for _, freq := range lfu.descFreq {
		if kvs, ok := lfu.freqMap[freq]; ok {
			items := make(map[string]interface{})
			for s, i := range kvs {
				if count >= num && num >= 0 {
					return results
				}
				items[s] = i

				results[freq] = items
				count++
			}
		}
	}

	return results
}

// GetBackEntriesKVS 返回后面数量的KV，-1返回所有
func (lfu *LFUCache) GetBackEntriesKVS(num int) map[int]map[string]interface{} {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	results := make(map[int]map[string]interface{})

	count := 0
	for _, freq := range lfu.ascFreq {
		if kvs, ok := lfu.freqMap[freq]; ok {
			if count >= num && num >= 0 {
				return results
			}
			results[freq] = kvs
			count++
		}
	}

	return results
}

// GetBackEntries 返回前面频率层级的所有KV，-1返回所有
func (lfu *LFUCache) GetBackEntries(num int) map[int]map[string]interface{} {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	results := make(map[int]map[string]interface{})

	count := 0
	for _, freq := range lfu.ascFreq {
		if kvs, ok := lfu.freqMap[freq]; ok {
			items := make(map[string]interface{})
			for s, i := range kvs {
				if count >= num && num >= 0 {
					return results
				}
				items[s] = i

				results[freq] = items
				count++
			}
		}
	}

	return results
}

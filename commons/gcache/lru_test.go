package gcache

import (
	"fmt"
	"sync"
	"testing"
)

func TestLRUCache(t *testing.T) {
	// 初始化LRU缓存
	lruCache := NewLRUCache(100)
	// 并发请求的数量

	concurrency := 100

	// 使用WaitGroup来等待所有goroutines完成
	var wg sync.WaitGroup
	wg.Add(concurrency)

	// 开启多个并发请求
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()

			key := fmt.Sprintf("key%d", id)
			// 向缓存中添加键值对
			lruCache.Put(key, id)

			// 从缓存中获取键值对
			val := lruCache.Get(key)
			if val == nil {
				fmt.Printf("Failed to retrieve value for key %s", key)
			}

			// 测试GetTopEntries和RemoveKey方法之间的竞争
			if id%2 == 0 {
				topEntries := lruCache.GetFrontEntries(10)
				fmt.Println("Top entries:", topEntries)
			} else {
				lruCache.RemoveKey(key)
			}
		}(i)
	}

	// 等待所有goroutines完成
	wg.Wait()

	// 检查缓存大小是否符合预期
	expectedSize := concurrency
	actualSize := lruCache.Size()
	if actualSize != expectedSize {
		fmt.Printf("Cache size mismatch: expected %d, got %d", expectedSize, actualSize)
	}
	fmt.Println("")
	fmt.Println(lruCache.GetFrontEntries(50))
	lruCache.RemoveFrontEntries(20)
	fmt.Println(lruCache.GetFrontEntries(50))
	fmt.Println(lruCache.Get("key66"))
	fmt.Println(lruCache.Get("key66"))

	fmt.Println(lruCache.GetFrontEntries(5))
	fmt.Println(lruCache.Get("key50"))
	fmt.Println(lruCache.GetBackEntries(3))
}

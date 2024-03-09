package utils

import (
	"fmt"
	"sync"
	"testing"
)

func TestLRUCache(t *testing.T) {
	// 初始化LRU缓存
	cache := NewLRUCache(100)
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
			cache.Put(key, id)

			// 从缓存中获取键值对
			val := cache.Get(key)
			if val == nil {
				fmt.Printf("Failed to retrieve value for key %s", key)
			}

			// 测试GetTopEntries和RemoveKey方法之间的竞争
			if id%2 == 0 {
				topEntries := cache.GetFrontEntries(10)
				fmt.Println("Top entries:", topEntries)
			} else {
				cache.RemoveKey(key)
			}
		}(i)
	}

	// 等待所有goroutines完成
	wg.Wait()

	// 检查缓存大小是否符合预期
	expectedSize := concurrency
	actualSize := cache.Size()
	if actualSize != expectedSize {
		fmt.Printf("Cache size mismatch: expected %d, got %d", expectedSize, actualSize)
	}
	fmt.Println("")
	fmt.Println(cache.GetFrontEntries(50))
	cache.RemoveFrontEntries(20)
	fmt.Println(cache.GetFrontEntries(50))
	fmt.Println(cache.Get("key66"))
	fmt.Println(cache.Get("key66"))

	fmt.Println(cache.GetFrontEntries(5))
	fmt.Println(cache.Get("key50"))
	fmt.Println(cache.GetBackEntries(3))
}

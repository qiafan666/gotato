package gcache

import (
	"fmt"
	"testing"
)

func TestLFUCache(t *testing.T) {
	cache := NewLFUCache(5)

	// Put一些值到缓存中
	cache.Put("1", 1)
	cache.Put("1", 1)
	cache.Put("1", 1)
	cache.Put("1", 1)
	cache.Put("2", 2)
	cache.Put("2", 2)
	cache.Put("3", 3)
	cache.Put("4", 4)
	cache.Put("4", 4)
	cache.Put("4", 4)
	cache.Put("5", 5)
	cache.Put("6", 6)
	cache.Put("7", 7)
	cache.Put("7", 7)
	cache.Remove("1")
	cache.Remove("5")
	cache.Put("8", 8)

	// 测试获取键值对
	fmt.Println("KeysAndValues:", cache.KeysAndValues())

	// 测试获取前3个键值对
	entries := cache.GetFrontEntries(3)
	fmt.Println(entries)
	fmt.Println("Front 3:", cache.GetFrontEntries(3))
	fmt.Println("FrontKVs 1:", cache.GetFrontFreqKVS(1))

	// 测试获取后2个键值对
	fmt.Println("Back 2:", cache.GetBackEntries(2))
	fmt.Println("BackKVs 1:", cache.GetBackEntriesKVS(1))

	// 测试移除键
	cache.Remove("2")
	fmt.Println("After removing key 2:", cache.KeysAndValues())

	// 测试清空缓存
	cache.Clear()
	fmt.Println("After clearing cache:", cache.KeysAndValues())

	// 再次放入一些值到缓存中
	cache.Put("A", "Apple")
	cache.Put("B", "Banana")
	cache.Put("B", "Banana")
	cache.Put("C", "Cat")
	cache.Put("C", "Cat")
	cache.Put("C", "Cat")
	cache.Put("C", "Cat")
	cache.Put("D", "Dog")
	cache.Put("D", "Dog")
	cache.Put("E", "Elephant")

	// 测试检查键是否存在
	fmt.Println("Contains key 'C':", cache.Contains("C"))

	// 测试获取缓存大小
	fmt.Println("Cache size:", cache.Size())

	// 测试获取所有键和值
	fmt.Println("FrontEntries 2:", cache.GetFrontEntries(2))
	fmt.Println("FrontEntries -1:", cache.GetFrontEntries(-1))
	fmt.Println("BackEntries 2:", cache.GetBackEntries(2))
	fmt.Println("BackEntries -1:", cache.GetBackEntries(-1))
	fmt.Println("FrontFreq 2:", cache.GetFrontFreq(2))
	fmt.Println("FrontFreq -1:", cache.GetFrontFreq(-1))
	fmt.Println("BackFreq 2:", cache.GetBackFreq(2))
	fmt.Println("BackFreq -1:", cache.GetBackFreq(-1))

	// 测试获取键值对
	fmt.Println("KeysAndValues:", cache.KeysAndValues())
	// 测试获取指定频率下的项
	fmt.Println("Items with frequency 1:", cache.FreqItems(1))
}

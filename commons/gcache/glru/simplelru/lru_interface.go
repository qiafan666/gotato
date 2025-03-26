package simplelru

type ILRUCache[K comparable, V any] interface {
	// Add 添加一个值到缓存中，如果发生了逐出操作则返回 true，并更新键的“最近使用”状态。
	Add(key K, value V) bool

	// Get 从缓存中返回键的值，并更新键的“最近使用”状态。#返回值, 是否找到
	Get(key K) (value V, ok bool)

	// Contains 检查一个键是否存在于缓存中，但不更新其最近使用状态。
	Contains(key K) (ok bool)

	// Peek 返回键的值，但不更新键的“最近使用”状态。
	Peek(key K) (value V, ok bool)

	// Remove 从缓存中移除一个键。
	Remove(key K) bool

	// RemoveOldest 从缓存中移除最旧的条目。
	RemoveOldest() (K, V, bool)

	// GetOldest 返回缓存中最旧的条目。#返回键, 值, 是否找到
	GetOldest() (K, V, bool)

	// Keys 返回缓存中所有键的切片，从最旧到最新排序。
	Keys() []K

	// Values 返回缓存中所有值的切片，从最旧到最新排序。
	Values() []V

	// Len 返回缓存中的项数。
	Len() int

	// Cap 返回缓存的容量。
	Cap() int

	// Purge 清除所有缓存条目。
	Purge()

	// Resize 调整缓存的大小，返回被逐出的项数。
	Resize(int) int
}

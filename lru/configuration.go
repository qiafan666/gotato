package lru

// Configuration[T] 表示LRU缓存的配置
type Configuration[T any] struct {
	maxSize        int64               // 最大缓存大小
	buckets        int                 // 桶数量
	itemsToPrune   int                 // 低内存时要修剪的项数
	deleteBuffer   int                 // 删除通道的缓冲区大小
	promoteBuffer  int                 // 提升通道的缓冲区大小
	getsPerPromote int32               // 每个键被提升前的获取次数
	tracking       bool                // 是否跟踪项
	onDelete       func(item *Item[T]) // 删除回调函数
}

// Configure 创建具有合理默认值的配置对象
// 可以使用此方法作为流式配置的起点：
// 例如：lru.New(lru.Configure().MaxSize(10000))
func Configure[T any]() *Configuration[T] {
	return &Configuration[T]{
		buckets:        16,
		itemsToPrune:   500,
		deleteBuffer:   1024,
		getsPerPromote: 3,
		promoteBuffer:  1024,
		maxSize:        5000,
		tracking:       false,
	}
}

// MaxSize 设置缓存的最大大小
// [5000]
func (c *Configuration[T]) MaxSize(max int64) *Configuration[T] {
	c.maxSize = max
	return c
}

// Buckets 将键哈希到％桶计数，以提供更大的并发性（每个设置都需要对桶进行写锁定）。
// 必须是2的幂（1、2、4、8、16等）
// [16]
func (c *Configuration[T]) Buckets(count uint32) *Configuration[T] {
	if count == 0 || !((count & (^count + 1)) == count) {
		count = 16
	}
	c.buckets = int(count)
	return c
}

// ItemsToPrune 当内存不足时修剪的项数
// [500]
func (c *Configuration[T]) ItemsToPrune(count uint32) *Configuration[T] {
	c.itemsToPrune = int(count)
	return c
}

// PromoteBuffer 用于应该提升的项的队列的大小。
// 如果队列填满，则跳过提升操作
// [1024]
func (c *Configuration[T]) PromoteBuffer(size uint32) *Configuration[T] {
	c.promoteBuffer = int(size)
	return c
}

// DeleteBuffer 用于应该删除的项的队列的大小。
// 如果队列填满，则 Delete() 调用将阻塞
func (c *Configuration[T]) DeleteBuffer(size uint32) *Configuration[T] {
	c.deleteBuffer = int(size)
	return c
}

// GetsPerPromote 对于读写比例较高的大缓存，通常不需要在每次 Get() 调用中都提升项。
// GetsPerPromote 指定了在提升之前键必须具有的 Get() 次数
// [3]
func (c *Configuration[T]) GetsPerPromote(count int32) *Configuration[T] {
	c.getsPerPromote = count
	return c
}

// Track 通常，缓存对于缓存的值如何使用是不可知的。对于典型的缓存用法，获取项并执行某些操作（将其写出）然后什么也不做是可以的。
// 但是，如果调用方要长时间保留对缓存项的引用，情况会变得混乱。具体来说，缓存可以在引用仍存在的情况下驱逐该项。
// 从技术上讲，这不是一个问题。但是，如果重新加载项目回到缓存中，您最终会得到表示相同数据的两个对象。这是一种浪费空间，并且可能导致奇怪的行为（身份映射类型是为了解决此问题）。
// 通过打开跟踪并使用缓存的 TrackingGet，缓存将不会驱逐您尚未调用 Release() 的项。这是一个简单的引用计数器。
func (c *Configuration[T]) Track() *Configuration[T] {
	c.tracking = true
	return c
}

// OnDelete 允许设置回调函数以对项删除做出反应。
// 这通常允许清理资源，例如调用 Close() 对缓存对象进行某种类型的拆除。
func (c *Configuration[T]) OnDelete(callback func(item *Item[T])) *Configuration[T] {
	c.onDelete = callback
	return c
}

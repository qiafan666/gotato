package lru

// 控制GC的消息
type controlGC struct {
	done chan struct{}
}

// 控制清除缓存的消息
type controlClear struct {
	done chan struct{}
}

// 控制停止缓存的消息
type controlStop struct {
}

// 控制获取缓存大小的消息
type controlGetSize struct {
	res chan int64
}

// 控制获取被删除的缓存项数量的消息
type controlGetDropped struct {
	res chan int
}

// 控制设置最大缓存大小的消息
type controlSetMaxSize struct {
	size int64
	done chan struct{}
}

// 控制同步更新的消息
type controlSyncUpdates struct {
	done chan struct{}
}

// 控制通道
type control chan interface{}

// 创建新的控制通道
func newControl() chan interface{} {
	return make(chan interface{}, 5)
}

// 强制执行GC。除了测试需要同步GC的情况外，不应该调用此函数。
// 这是一个控制命令。
func (c control) GC() {
	done := make(chan struct{})
	c <- controlGC{done: done}
	<-done
}

// 发送停止信号给工作线程。工作线程将在接收到最后一条消息后的5秒内关闭。
// 在调用 Stop 后不应再使用缓存，但并发执行的请求应该正确完成执行。
// 这是一个控制命令。
func (c control) Stop() {
	c.SyncUpdates()
	c <- controlStop{}
}

// 清除缓存
// 这是一个控制命令。
func (c control) Clear() {
	done := make(chan struct{})
	c <- controlClear{done: done}
	<-done
}

// 获取缓存的大小。这是一个O(1)的调用，但由工作线程处理。
// 它被设计为定期用于指标，或从测试中调用。
// 这是一个控制命令。
func (c control) GetSize() int64 {
	res := make(chan int64)
	c <- controlGetSize{res: res}
	return <-res
}

// 获取由于内存压力而从缓存中删除的项目数量，自上次调用 GetDropped 后的数量。
// 这是一个控制命令。
func (c control) GetDropped() int {
	res := make(chan int)
	c <- controlGetDropped{res: res}
	return <-res
}

// 设置新的最大缓存大小。如果新的最大大小小于缓存的当前大小，则可能会触发GC。
// 这是一个控制命令。
func (c control) SetMaxSize(size int64) {
	done := make(chan struct{})
	c <- controlSetMaxSize{size: size, done: done}
	<-done
}

// SyncUpdates 等待直到缓存完成当前goroutine已经完成的任何操作的异步状态更新。
//
// 为了效率，LRU行为的实现部分受到工作线程的管理，该工作线程异步更新其内部数据结构。
// 这意味着缓存的状态（例如LRU项目的驱逐）仅具有最终一致性；不能保证在Get或Set调用返回之前发生。
// 大多数情况下，应用程序代码不会关心这一点，但特别是在测试场景中，您可能希望能够知道工作线程何时赶上。
//
// 这仅适用于之前由同一goroutine调用的缓存方法现在调用 SyncUpdates。如果其他goroutine同时使用缓存，则无法知道 SyncUpdates 返回时它们是否仍有待处理的状态更新。
// 这是一个控制命令。
func (c control) SyncUpdates() {
	done := make(chan struct{})
	c <- controlSyncUpdates{done: done}
	<-done
}

package gcommon

import "container/heap"

// GenericHeap 实现了 heap.Interface 接口，用于自定义排序。
type GenericHeap[T any] struct {
	data []T
	less func(a, b T) bool // 用于自定义排序的比较函数
}

// Len 返回堆中元素的数量。
func (h *GenericHeap[T]) Len() int { return len(h.data) }

// Less 根据自定义的比较函数比较两个元素。
func (h *GenericHeap[T]) Less(i, j int) bool {
	return h.less(h.data[i], h.data[j])
}

// Swap 交换堆中两个元素的位置。
func (h *GenericHeap[T]) Swap(i, j int) {
	h.data[i], h.data[j] = h.data[j], h.data[i]
}

// Push 向堆中添加一个元素。
func (h *GenericHeap[T]) Push(x interface{}) {
	h.data = append(h.data, x.(T))
}

// Pop 移除并返回最小的元素（根据 Less 定义）。
func (h *GenericHeap[T]) Pop() interface{} {
	old := h.data
	n := len(old)
	item := old[n-1]
	h.data = old[0 : n-1]
	return item
}

// HeapSort 使用自定义的比较函数对任意类型的切片进行排序。
func HeapSort[T any](data []T, less func(a, b T) bool) []T {
	// 使用提供的数据和比较函数初始化泛型堆。
	h := &GenericHeap[T]{data: data, less: less}
	heap.Init(h)

	// 从堆中逐个弹出元素，获取排序后的结果。
	sorted := make([]T, 0, len(data))
	for h.Len() > 0 {
		sorted = append(sorted, heap.Pop(h).(T))
	}

	return sorted
}

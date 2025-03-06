package gcommon

import "container/heap"

// genericHeap 实现了 heap.Interface 接口，用于自定义排序。
type genericHeap[T any] struct {
	data []T
	less func(a, b T) bool // 用于自定义排序的比较函数
}

// Len 返回堆中元素的数量。
func (h *genericHeap[T]) Len() int { return len(h.data) }

// Less 根据自定义的比较函数比较两个元素。
func (h *genericHeap[T]) Less(i, j int) bool {
	return h.less(h.data[i], h.data[j])
}

// Swap 交换堆中两个元素的位置。
func (h *genericHeap[T]) Swap(i, j int) {
	h.data[i], h.data[j] = h.data[j], h.data[i]
}

// Push 向堆中添加一个元素。
func (h *genericHeap[T]) Push(x interface{}) {
	h.data = append(h.data, x.(T))
}

// Pop 移除并返回最小的元素（根据 Less 定义）。
func (h *genericHeap[T]) Pop() interface{} {
	old := h.data
	n := len(old)
	item := old[n-1]
	h.data = old[0 : n-1]
	return item
}

// HeapSort 使用自定义的比较函数对任意类型的切片进行排序。
func HeapSort[T any](data []T, less func(a, b T) bool) []T {
	// 使用提供的数据和比较函数初始化泛型堆。
	h := &genericHeap[T]{data: data, less: less}
	heap.Init(h)

	// 从堆中逐个弹出元素，获取排序后的结果。
	sorted := make([]T, 0, len(data))
	for h.Len() > 0 {
		sorted = append(sorted, heap.Pop(h).(T))
	}

	return sorted
}

// HeapSortFilter 对数据进行排序并按过滤条件筛选元素。
func HeapSortFilter[T any](data []T, less func(a, b T) bool, filter func(T) bool) []T {
	// 使用提供的数据和比较函数初始化泛型堆。
	h := &genericHeap[T]{data: data, less: less}
	heap.Init(h)

	// 从堆中逐个弹出元素，获取排序后的结果。
	sorted := make([]T, 0, len(data))
	for h.Len() > 0 {
		item := heap.Pop(h).(T)
		// 如果 filter 函数不为空且当前元素符合过滤条件，则将其添加到结果中。
		if filter == nil || filter(item) {
			sorted = append(sorted, item)
		}
	}

	return sorted
}

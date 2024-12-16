package gcommon

import (
	"github.com/qiafan666/gotato/commons/gcast"
	mathRand "math/rand"
	"reflect"
	"sort"
	"strings"
	"sync"
)

// SliceContain 返回切片是否包含指定元素
func SliceContain[T ~[]E, E comparable](list T, elem ...E) bool {
	if len(elem) == 0 {
		return false
	}
	for _, v := range elem {
		if contain(list, v) {
			return true
		}
	}
	return false
}

func contain[T ~[]E, E comparable](list T, elem E) bool {
	if len(list) == 0 {
		return false
	}
	for _, v := range list {
		if v == elem {
			return true
		}
	}
	return false
}

// SliceContains 返回A集合是否包含B集合里的任意元素
func SliceContains[T ~[]E, E comparable](list1 T, list2 T) bool {
	if len(list1) == 0 || len(list2) == 0 {
		return false
	}
	m := make(map[E]struct{}, len(list2))
	for _, v := range list2 {
		m[v] = struct{}{}
	}
	for _, v := range list1 {
		if _, ok := m[v]; ok {
			return true
		}
	}
	return false
}

// Slice2String 将切片转换为分隔符分隔的字符串
func Slice2String[T any](list []T, sep string) string {
	if len(list) == 0 {
		return ""
	}
	convert := SliceConvert(list, func(e T) string {
		return gcast.ToString(e)
	})
	return strings.Join(convert, sep)
}

// String2Slice 将字符串转换为切片
func String2Slice(str string, sep string) []string {
	if str == "" {
		return nil
	}
	return strings.Split(str, sep)
}

// SliceClone 深拷贝切片
func SliceClone[S ~[]E, E any](s S) S {
	return append(s[:0:0], s...)
}

// SliceUniq 集合去重
func SliceUniq[T ~[]E, E comparable](list T) T {
	if len(list) == 0 {
		return list
	}

	ret := make(T, 0, len(list))
	m := make(map[E]struct{}, len(list))
	for _, v := range list {
		if _, ok := m[v]; !ok {
			ret = append(ret, v)
			m[v] = struct{}{}
		}
	}
	return ret
}

// SliceDiff 返回两个集合之间的差异
func SliceDiff[T ~[]E, E comparable](list1 T, list2 T) (ret1 T, ret2 T) {
	m1 := map[E]struct{}{}
	m2 := map[E]struct{}{}
	for _, v := range list1 {
		m1[v] = struct{}{}
	}
	for _, v := range list2 {
		m2[v] = struct{}{}
	}

	ret1 = make(T, 0)
	ret2 = make(T, 0)
	for _, v := range list1 {
		if _, ok := m2[v]; !ok {
			ret1 = append(ret1, v)
		}
	}
	for _, v := range list2 {
		if _, ok := m1[v]; !ok {
			ret2 = append(ret2, v)
		}
	}
	return ret1, ret2
}

// SliceWithout 返回不包括所有给定值的切片
func SliceWithout[T ~[]E, E comparable](list T, exclude ...E) T {
	if len(list) == 0 {
		return list
	}

	m := make(map[E]struct{}, len(exclude))
	for _, v := range exclude {
		m[v] = struct{}{}
	}

	ret := make(T, 0, len(list))
	for _, v := range list {
		if _, ok := m[v]; !ok {
			ret = append(ret, v)
		}
	}
	return ret
}

// SliceIntersect 返回两个集合的交集
func SliceIntersect[T ~[]E, E comparable](list1 T, list2 T) T {
	m := make(map[E]struct{})
	for _, v := range list1 {
		m[v] = struct{}{}
	}

	ret := make(T, 0)
	for _, v := range list2 {
		if _, ok := m[v]; ok {
			ret = append(ret, v)
		}
	}
	return ret
}

// SliceUnion 返回两个集合的并集
func SliceUnion[T ~[]E, E comparable](lists ...T) T {
	ret := make(T, 0)
	m := make(map[E]struct{})
	for _, list := range lists {
		for _, v := range list {
			if _, ok := m[v]; !ok {
				ret = append(ret, v)
				m[v] = struct{}{}
			}
		}
	}
	return ret
}

// SliceRand 返回一个指定随机挑选个数的切片
// 若 n == -1 or n >= len(list)，则返回打乱的切片
func SliceRand[T ~[]E, E any](list T, n int) T {
	if n == 0 || n < -1 {
		return nil
	}

	count := len(list)
	ret := make(T, count)
	copy(ret, list)
	mathRand.Shuffle(count, func(i, j int) {
		ret[i], ret[j] = ret[j], ret[i]
	})
	if n == -1 || n >= count {
		return ret
	}
	return ret[:n]
}

// SlicePinTop 置顶集合中的一个元素
func SlicePinTop[T any](list []T, index int) {
	if index <= 0 || index >= len(list) {
		return
	}
	for i := index; i > 0; i-- {
		list[i], list[i-1] = list[i-1], list[i]
	}
}

// SlicePinTopF 置顶集合中满足条件的一个元素
func SlicePinTopF[T any](list []T, fn func(v T) bool) {
	index := 0
	for i, v := range list {
		if fn(v) {
			index = i
			break
		}
	}
	for i := index; i > 0; i-- {
		list[i], list[i-1] = list[i-1], list[i]
	}
}

// SliceAppendUnique 数组中是否包含某个元素,没有就追加,有就返回原数组
func SliceAppendUnique[T any](ts []T, t T) []T {
	for _, v := range ts {
		if reflect.DeepEqual(v, t) {
			return ts
		}
	}
	ts = append(ts, t)
	return ts
}

// SliceDelete 使用泛型函数来删除切片中的某个元素
func SliceDelete[T any](ts []T, t T) []T {
	result := make([]T, 0, len(ts))
	for _, v := range ts {
		if !reflect.DeepEqual(v, t) {
			result = append(result, v)
		}
	}
	return result
}

// SliceDeleteF 使用泛型函数来删除切片中的某个元素
func SliceDeleteF[T any](ts []T, fn func(t T) bool) []T {
	result := make([]T, 0, len(ts))
	for _, v := range ts {
		if !fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// SliceIndex 返回元素在切片中的索引，不存在返回-1
func SliceIndex[T any](list []T, elem T) int {
	for i, v := range list {
		if reflect.DeepEqual(v, elem) {
			return i
		}
	}
	return -1
}

// SliceIndexF 返回元素在切片中的索引
func SliceIndexF[T any](list []T, fn func(elem T) bool) []int {
	indexes := make([]int, 0)
	for i, v := range list {
		if fn(v) {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

// SliceDeleteIndex 高效地删除指定索引的多个元素
func SliceDeleteIndex[T any](list []T, indexes ...int) []T {
	if len(indexes) == 0 {
		return list
	}
	sort.Ints(indexes)
	// 创建一个 map 以 O(1) 复杂度检查索引是否需要删除
	indexMap := make(map[int]struct{}, len(indexes))
	for _, idx := range indexes {
		indexMap[idx] = struct{}{}
	}

	// 构建结果切片，跳过 map 中的索引
	result := make([]T, 0, len(list)-len(indexes))
	for i, v := range list {
		if _, shouldDelete := indexMap[i]; !shouldDelete {
			result = append(result, v)
		}
	}

	return result
}

// sliceToMapOkAny 切片转映射（自定义类型，过滤器）
func sliceToMapOkAny[E any, K comparable, V any](es []E, fn func(e E) (K, V, bool)) map[K]V {
	kv := make(map[K]V)
	for i := 0; i < len(es); i++ {
		t := es[i]
		if k, v, ok := fn(t); ok {
			kv[k] = v
		}
	}
	return kv
}

// SliceToMapAny 切片转映射（自定义类型）
func SliceToMapAny[E any, K comparable, V any](es []E, fn func(e E) (K, V)) map[K]V {
	return sliceToMapOkAny(es, func(e E) (K, V, bool) {
		k, v := fn(e)
		return k, v, true
	})
}

// SliceToMap 只处理键，值和数组元素相同
func SliceToMap[E any, K comparable](es []E, fn func(e E) K) map[K]E {
	return sliceToMapOkAny(es, func(e E) (K, E, bool) {
		k := fn(e)
		return k, e, true
	})
}

func SliceSetAny[E any, K comparable](es []E, fn func(e E) K) map[K]struct{} {
	return SliceToMapAny(es, func(e E) (K, struct{}) {
		return fn(e), struct{}{}
	})
}

// SliceToNilMap 切片转换为空对象的map
func SliceToNilMap[E comparable](es []E) map[E]struct{} {
	return SliceSetAny(es, func(e E) E {
		return e
	})
}

func SliceFilter[E, T any](es []E, fn func(e E) (T, bool)) []T {
	rs := make([]T, 0, len(es))
	for i := 0; i < len(es); i++ {
		e := es[i]
		if t, ok := fn(e); ok {
			rs = append(rs, t)
		}
	}
	return rs
}

// SliceConvert 转换切片
func SliceConvert[E any, T any](es []E, fn func(e E) T) []T {
	v := make([]T, len(es))
	for i := 0; i < len(es); i++ {
		v[i] = fn(es[i])
	}
	return v
}

// SlicePaginate 分页
func SlicePaginate[E any](es []E, pageNumber int, pageSize int) []E {
	if pageNumber < 0 {
		pageNumber = 0
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	start := min(pageNumber*pageSize, len(es))
	end := min(start+pageSize, len(es))
	return es[start:end]
}

// If 如果isa为true，则返回a，否则返回b
func If[T any](isa bool, a, b T) T {
	if isa {
		return a
	}
	return b
}

func ToPtr[T any](t T) *T {
	return &t
}

// SliceBatch 批量处理切片 numWorkers为0或1时，串行处理；numWorkers大于1时，并行处理
func SliceBatch[T any, V any](ts []T, fn func(T) V, numWorkers ...int) []V {
	if ts == nil {
		return nil
	}

	if len(numWorkers) <= 1 {
		res := make([]V, 0, len(ts))
		for i := range ts {
			res = append(res, fn(ts[i]))
		}
		return res
	} else {
		var wg sync.WaitGroup
		chunkSize := (len(ts) + numWorkers[0] - 1) / numWorkers[0]

		// 通道用于收集结果
		resultChan := make(chan V, len(ts))

		for i := 0; i < numWorkers[0]; i++ {
			wg.Add(1)
			go func(start int) {
				defer wg.Done()
				end := start + chunkSize
				if end > len(ts) {
					end = len(ts)
				}
				for j := start; j < end; j++ {
					resultChan <- fn(ts[j])
				}
			}(i * chunkSize)
		}

		// 等待所有协程完成并关闭通道
		go func() {
			wg.Wait()
			close(resultChan)
		}()

		// 收集所有结果
		res := make([]V, 0, len(ts))
		for r := range resultChan {
			res = append(res, r)
		}

		return res
	}
}

// Equal 比较两个切片是否相等，元素顺序相关
func Equal[E comparable](a []E, b []E) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// SliceSort 排序切片
func SliceSort[E any](es []E, fn func(a, b E) bool) {
	sort.Sort(&sortSlice[E]{
		ts: es,
		fn: fn,
	})
}

type sortSlice[E any] struct {
	ts []E
	fn func(a, b E) bool
}

func (o *sortSlice[E]) Len() int {
	return len(o.ts)
}

func (o *sortSlice[E]) Less(i, j int) bool {
	return o.fn(o.ts[i], o.ts[j])
}

func (o *sortSlice[E]) Swap(i, j int) {
	o.ts[i], o.ts[j] = o.ts[j], o.ts[i]
}

// SliceReverse 反转切片
func SliceReverse[E any](es []E) {
	for i, j := 0, len(es)-1; i < j; i, j = i+1, j-1 {
		es[i], es[j] = es[j], es[i]
	}
}

// SliceForEach 遍历切片
func SliceForEach[E any](es []E, fn func(e E)) {
	for i := 0; i < len(es); i++ {
		fn(es[i])
	}
}

// SliceForEachReverse 反向遍历切片
func SliceForEachReverse[E any](es []E, fn func(e E)) {
	for i := len(es) - 1; i >= 0; i-- {
		fn(es[i])
	}
}

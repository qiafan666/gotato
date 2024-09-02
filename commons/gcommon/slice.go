package gcommon

import (
	mathRand "math/rand"
	"reflect"
	"sort"
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

// contain 返回切片是否包含指定元素
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

// SliceDeleteIndex 删除多个元素的索引
func SliceDeleteIndex[T any](list []T, indexes ...int) []T {
	if len(indexes) == 0 {
		return list
	}
	if len(list) == 0 {
		return list
	}
	for _, index := range indexes {
		if index < 0 || index >= len(list) {
			continue
		}
		list = append(list[:index], list[index+1:]...)
	}
	return list
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

// SliceRemove 使用泛型函数来删除切片中的某个元素
func SliceRemove[T any](ts []T, t T) []T {
	for i, v := range ts {
		if reflect.DeepEqual(v, t) {
			return append(ts[:i], ts[i+1:]...)
		}
	}
	return ts // 如果未找到匹配的元素，则返回原始切片
}

// SliceToMapOkAny 切片转映射（自定义类型，过滤器）
func SliceToMapOkAny[E any, K comparable, V any](es []E, fn func(e E) (K, V, bool)) map[K]V {
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
	return SliceToMapOkAny(es, func(e E) (K, V, bool) {
		k, v := fn(e)
		return k, v, true
	})
}

// SliceToMap slice to map
func SliceToMap[E any, K comparable](es []E, fn func(e E) K) map[K]E {
	return SliceToMapOkAny(es, func(e E) (K, E, bool) {
		k := fn(e)
		return k, e, true
	})
}

// SliceSetAny slice to map[K]struct{}
func SliceSetAny[E any, K comparable](es []E, fn func(e E) K) map[K]struct{} {
	return SliceToMapAny(es, func(e E) (K, struct{}) {
		return fn(e), struct{}{}
	})
}

func Filter[E, T any](es []E, fn func(e E) (T, bool)) []T {
	rs := make([]T, 0, len(es))
	for i := 0; i < len(es); i++ {
		e := es[i]
		if t, ok := fn(e); ok {
			rs = append(rs, t)
		}
	}
	return rs
}

// Slice 转换切片
func Slice[E any, T any](es []E, fn func(e E) T) []T {
	v := make([]T, len(es))
	for i := 0; i < len(es); i++ {
		v[i] = fn(es[i])
	}
	return v
}

// SliceSet slice to map[E]struct{}
func SliceSet[E comparable](es []E) map[E]struct{} {
	return SliceSetAny(es, func(e E) E {
		return e
	})
}

func paginate[E any](es []E, pageNumber int, pageSize int) []E {
	if pageNumber <= 0 {
		return []E{}
	}
	if pageSize <= 0 {
		return []E{}
	}
	start := (pageSize - 1) * pageSize
	end := start + pageSize
	if start >= len(es) {
		return []E{}
	}
	if end > len(es) {
		end = len(es)
	}
	return es[start:end]
}

func SlicePaginate[E any](es []E, pageNumber int, pageSize int) []E {
	return paginate(es, pageNumber, pageSize)
}

// SortAny custom sort method
func SortAny[E any](es []E, fn func(a, b E) bool) {
	sort.Sort(&sortSlice[E]{
		ts: es,
		fn: fn,
	})
}

// If true -> a, false -> b
func If[T any](isa bool, a, b T) T {
	if isa {
		return a
	}
	return b
}

func ToPtr[T any](t T) *T {
	return &t
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

// SliceBatch 批量处理切片
func SliceBatch[T any, V any](ts []T, fn func(T) V) []V {
	if ts == nil {
		return nil
	}
	res := make([]V, 0, len(ts))
	for i := range ts {
		res = append(res, fn(ts[i]))
	}
	return res
}

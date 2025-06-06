package gcommon

import (
	"fmt"
	"github.com/qiafan666/gotato/commons/gcast"
	"github.com/qiafan666/gotato/commons/gconc"
	"github.com/qiafan666/gotato/commons/gerr"
	"sort"
	"strings"
	"sync"
)

// MapMerge 函数用于合并多个map
func MapMerge[K comparable, V any](maps ...map[K]V) map[K]V {
	// 参数校验
	if len(maps) == 0 {
		return nil
	}

	mergedMap := make(map[K]V)
	for _, m := range maps {
		if m == nil || len(m) == 0 {
			continue // 跳过 nil map
		}
		for k, v := range m {
			mergedMap[k] = v
		}
	}
	return mergedMap
}

// MapMergeE 函数用于合并多个map,相同key报错
func MapMergeE[K comparable, V any](maps ...map[K]V) (map[K]V, error) {
	// 参数校验
	if len(maps) == 0 {
		return nil, nil
	}

	mergedMap := make(map[K]V)
	for _, m := range maps {
		if m == nil || len(m) == 0 {
			continue // 跳过 nil map
		}
		for k, v := range m {
			if _, exists := mergedMap[k]; exists {
				return nil, gerr.New("duplicate key", "key", k)
			}
			mergedMap[k] = v
		}
	}
	return mergedMap, nil
}

// MapMergeUnique 函数用于合并多个map, 并确保每个 key 只出现一次（保留第一个出现的键值对）
func MapMergeUnique[K comparable, V any](maps ...map[K]V) map[K]V {
	// 参数校验
	if len(maps) == 0 {
		return nil
	}

	mergedMap := make(map[K]V)
	for _, m := range maps {
		if m == nil || len(m) == 0 {
			continue // 跳过 nil map
		}
		for k, v := range m {
			if _, exists := mergedMap[k]; !exists {
				mergedMap[k] = v
			}
		}
	}
	return mergedMap
}

// MapMergeConcurrent 并发安全的合并多个map
func MapMergeConcurrent[K comparable, V any](maps ...map[K]V) map[K]V {
	// 参数校验
	if len(maps) == 0 {
		return nil
	}

	mergedMap := make(map[K]V)
	var mu sync.Mutex
	wg := gconc.NewWaitGroup()

	for _, m := range maps {
		if m == nil {
			continue // 跳过 nil map
		}
		wg.Go(func() {
			for k, v := range m {
				mu.Lock()
				mergedMap[k] = v
				mu.Unlock()
			}
		})
	}

	wg.Wait()
	return mergedMap
}

// MapKeys 获取map的key列表
func MapKeys[K comparable, V any](kv map[K]V) []K {
	ks := make([]K, 0, len(kv))
	for k := range kv {
		ks = append(ks, k)
	}
	return ks
}

// MapValues 获取map的value列表
func MapValues[K comparable, V any](kv map[K]V) []V {
	vs := make([]V, 0, len(kv))
	for k := range kv {
		vs = append(vs, kv[k])
	}
	return vs
}

// MapClone 深拷贝map
func MapClone[K comparable, V any](m map[K]V) map[K]V {
	// 创建一个新的 map
	cloned := make(map[K]V, len(m))
	for key, value := range m {
		cloned[key] = value // 复制每个键值对
	}
	return cloned
}

// MapSortKey 排序map的key
func MapSortKey[K comparable, V any](m map[K]V, cmp func(a, b K) bool) []K {
	keys := MapKeys(m)
	SliceSort(keys, cmp)
	return keys // 返回排序后的键切片
}

// MapSortValue 排序map的value
func MapSortValue[K comparable, V any](m map[K]V, cmp func(a, b V) bool) []V {
	values := MapValues(m)
	SliceSort(values, cmp)
	return values // 返回排序后的值切片
}

// MapSortUrl 根据字符顺序排序map并拼接成字符串
func MapSortUrl(m map[string]any) string {
	if len(m) == 0 {
		return ""
	}

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var queryString strings.Builder
	for i, key := range keys {
		queryString.WriteString(fmt.Sprintf("%s=%s", key, gcast.ToString(m[key])))
		if i < len(keys)-1 {
			queryString.WriteString("&")
		}
	}
	return queryString.String()
}

// MapKv2Slice 将map的键值对转换成切片
func MapKv2Slice[K comparable, V any](m map[K]V) []any {
	var slice []any
	for k, v := range m {
		slice = append(slice, k, v)
	}
	return slice
}

// MapK2Slice 将map的键转换成切片
func MapK2Slice[K comparable, V any](m map[K]V) []K {
	var slice []K
	for k := range m {
		slice = append(slice, k)
	}
	return slice
}

// MapV2Slice 将map的值转换成切片
func MapV2Slice[K comparable, V any](m map[K]V) []V {
	var slice []V
	for _, v := range m {
		slice = append(slice, v)
	}
	return slice
}

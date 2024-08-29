package gmap

import "sync"

// MergeMaps 函数用于合并多个map
func MergeMaps[K comparable, V any](maps ...map[K]V) map[K]V {
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

// MergeMapsUnique 函数用于合并多个map, 并确保每个 key 只出现一次（保留第一个出现的键值对）
func MergeMapsUnique[K comparable, V any](maps ...map[K]V) map[K]V {
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

// MergeMapsConcurrent 并发安全的合并多个map
func MergeMapsConcurrent[K comparable, V any](maps ...map[K]V) map[K]V {
	// 参数校验
	if len(maps) == 0 {
		return nil
	}

	mergedMap := make(map[K]V)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, m := range maps {
		if m == nil {
			continue // 跳过 nil map
		}
		wg.Add(1)
		go func(m map[K]V) {
			defer wg.Done()
			for k, v := range m {
				mu.Lock()
				mergedMap[k] = v
				mu.Unlock()
			}
		}(m)
	}

	wg.Wait()
	return mergedMap
}

// Keys get map keys
func Keys[K comparable, V any](kv map[K]V) []K {
	ks := make([]K, 0, len(kv))
	for k := range kv {
		ks = append(ks, k)
	}
	return ks
}

// Values get map values
func Values[K comparable, V any](kv map[K]V) []V {
	vs := make([]V, 0, len(kv))
	for k := range kv {
		vs = append(vs, kv[k])
	}
	return vs
}

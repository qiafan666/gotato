package gmap

import "sync"

// MergeMaps 函数用于合并多个同属性的 map
func MergeMaps[K comparable, V any](maps ...map[K]V) map[K]V {
	// 参数校验
	if len(maps) == 0 {
		return nil
	}

	mergedMap := make(map[K]V)
	for _, m := range maps {
		if m == nil {
			continue // 跳过 nil map
		}
		for k, v := range m {
			mergedMap[k] = v
		}
	}
	return mergedMap
}

// MergeMapsUnique 函数用于合并多个同属性的 map, 并确保每个 key 只出现一次（保留第一个出现的键值对）
func MergeMapsUnique[K comparable, V any](maps ...map[K]V) map[K]V {
	// 参数校验
	if len(maps) == 0 {
		return nil
	}

	mergedMap := make(map[K]V)
	for _, m := range maps {
		if m == nil {
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

// MergeMapsConcurrent 并发安全的合并多个同属性的 map
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

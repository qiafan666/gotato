package utils

import "sync"

// 获取
func SyncMapGet[T any](m *sync.Map, key any) (value T, ok bool) {
	load, ok := m.Load(key)
	if ok {
		value = load.(T)
	}
	return value, ok
}

// 设置
func SyncMapSet(m *sync.Map, key any, value any) {
	m.Store(key, value)
}

// 遍历,返回false终止遍历
func SyncMapRange(m *sync.Map, f func(key, value any) bool) {
	m.Range(f)
}

// 删除
func SyncMapDelete(m *sync.Map, key any) {
	m.Delete(key)
}

package utils

import (
	"fmt"
	"sync"
	"testing"
)

func TestSyncMap(t *testing.T) {

	syncMap := sync.Map{}
	SyncMapSet(&syncMap, "test1", "test1")
	SyncMapSet(&syncMap, "test2", "test2")
	SyncMapSet(&syncMap, "test3", "test3")

	value, ok := SyncMapGet[string](&syncMap, "test1")
	if ok {
		fmt.Println(value)
	}

	var valueSlice []string
	SyncMapRange(&syncMap, func(key, value any) bool {
		valueSlice = append(valueSlice, value.(string))
		return true
	})
	fmt.Println(valueSlice)

	SyncMapDelete(&syncMap, "test2")

	SyncMapRange(&syncMap, func(key, value any) bool {
		valueSlice = append(valueSlice, value.(string))
		return true
	})

	fmt.Println(valueSlice)
}

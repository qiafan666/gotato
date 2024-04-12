package utils

import (
	"sync"
	"time"
)

func init() {
	go func() {
		tick := time.Tick(time.Second)
		for {
			<-tick
			clearCache()
		}
	}()
}

var syncMaps sync.Map

type container struct {
	Expiration int64
	Value      interface{}
}

func GetCache[T any](key string) (T, bool) {

	var zero T
	i, exist := syncMaps.Load(key)
	if !exist {
		return zero, false
	}
	t := i.(container)

	if t.Expiration != 0 {

		if time.Now().Unix() > t.Expiration {
			return zero, false
		}
	}

	return t.Value.(T), true
}

func SetCache(key string, value interface{}) {
	syncMaps.Store(key, container{
		Expiration: 0,
		Value:      value,
	})

}

// SetCacheExpire 设置缓存,过期时间以秒为单位 eg:int64(time.Second * 10)
func SetCacheExpire(key string, value interface{}, expire int64) {
	syncMaps.Store(key, container{
		Expiration: time.Now().Unix() + expire,
		Value:      value,
	})
}

func clearCache() {
	syncMaps.Range(func(key, value interface{}) bool {
		t := value.(container)
		if t.Expiration != 0 {
			if time.Now().Unix() > t.Expiration {
				syncMaps.Delete(key)
			}
		}
		return true
	})
}

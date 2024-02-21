package middleware

import "sync"

var ignoreRequestMap sync.Map

func RegisterIgnoreRequest(path ...string) {
	for _, v := range path {
		ignoreRequestMap.Store(v, true)
	}
}

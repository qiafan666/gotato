package middleware

import (
	"regexp"
	"strings"
	"sync"
)

var ignoreRequestMap sync.Map

// RegisterIgnoreRequest 忽略打印当前路径的接口日志
func RegisterIgnoreRequest(paths ...string) {
	for _, path := range paths {
		// 如果路径中包含通配符 /*，则将其替换为正则表达式中的通配符 .*
		currentPath := path
		if strings.Contains(path, "/*") {
			currentPath = strings.Replace(path, "/*", "/.*", -1)
		}

		if _, exist := ignoreRequestMap.Load(currentPath); !exist {
			ignoreRequestMap.Store(path, true)
		}
	}
}

// IsIgnoredRequest 判断请求路径是否应该被忽略
func IsIgnoredRequest(requestPath string) bool {
	var isIgnored bool
	ignoreRequestMap.Range(func(key, value interface{}) bool {
		pathPattern := key.(string)
		// 使用正则表达式匹配请求路径
		matched, err := regexp.MatchString(pathPattern, requestPath)
		if err == nil && matched {
			isIgnored = true
			return false // 停止 Range 循环
		}
		return true // 继续 Range 循环
	})
	return isIgnored
}

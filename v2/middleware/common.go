package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"regexp"
	"strings"
	"sync"
)

var ignoreRequestMap sync.Map

// RegisterIgnoreRequest 忽略打印当前路径的接口日志
func RegisterIgnoreRequest(paths ...string) {
	for _, path := range paths {
		// 如果路径中包含通配符 /*，则将其替换为正则表达式中的通配符 .*
		if strings.Contains(path, "/*") {
			regexPath := strings.Replace(path, "/*", "/.*", -1)
			ignoreRequestMap.Store(regexPath, true)
		} else {
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

// CustomResponseWriter 自定义响应写入器结构体
type CustomResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 实现了 io.Writer 接口中的 Write 方法，用于写入字节切片，并记录到响应体中
func (w *CustomResponseWriter) Write(b []byte) (int, error) {
	// 将字节切片写入到响应体中
	n, err := w.body.Write(b)
	if err != nil {
		return n, err
	}
	// 写入响应体
	return w.ResponseWriter.Write(b)
}

// WriteString 实现了 WriteString 方法，用于写入字符串，并记录到响应体中
func (w *CustomResponseWriter) WriteString(s string) (int, error) {
	// 将字符串写入到响应体中
	n, err := w.body.WriteString(s)
	if err != nil {
		return n, err
	}
	// 写入响应体
	return w.ResponseWriter.WriteString(s)
}

package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"sync"
)

var ignoreRequestMap sync.Map

// RegisterIgnoreRequest 忽略打印当前路径的接口日志
func RegisterIgnoreRequest(path ...string) {
	for _, v := range path {
		ignoreRequestMap.Store(v, true)
	}
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

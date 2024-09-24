package gcommon

import (
	"bytes"
	"github.com/qiafan666/gotato/commons/gcast"
	"strings"
)

// Buffer 内嵌bytes.Buffer，支持连写
type Buffer struct {
	*bytes.Buffer
}

// NewBuffer 创建Buffer
func NewBuffer() *Buffer {
	return &Buffer{Buffer: new(bytes.Buffer)}
}

// Append 追加数据对象
func (b *Buffer) Append(i ...any) *Buffer {
	if len(i) == 0 {
		return b
	}

	var builder strings.Builder
	for _, v := range i {
		builder.WriteString(gcast.ToString(v))
	}
	b.WriteString(builder.String())

	return b
}

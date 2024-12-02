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

// AppendStr 追加字符串
func AppendStr(i ...any) *Buffer {
	return NewBuffer().Append(i...)
}

// AppendSplit 追加字符串，并用指定分隔符分隔
func AppendSplit(sep string, strings ...any) *Buffer {
	if len(strings) == 0 {
		return NewBuffer()
	}

	buffer := NewBuffer()
	for i := 0; i < len(strings); i++ {
		if i == len(strings)-1 {
			buffer.Append(strings[i])
		} else {
			buffer.Append(strings[i], sep)
		}
	}
	return buffer
}

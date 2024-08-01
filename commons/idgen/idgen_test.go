package idgen

import (
	"fmt"
	"testing"
	"time"
)

func TestGenerateID(t *testing.T) {

	sid := int32(4095)
	generator := NewIDGenerator(sid)
	id := generator.NewID()
	sid, createTime := ParseID(id)
	fmt.Println(sid, createTime) // 1 158992640
	fmt.Println(IsServerIDValid(sid))

	sid1 := int32(4094)
	idGenerator := NewIDGenerator(sid1)
	id1 := idGenerator.NewID()
	fmt.Println(id1) // 7398101349412372480
	fmt.Println(sid1, createTime)
	fmt.Println(IsServerIDValid(sid))
}

func BenchmarkGenerateID(b *testing.B) {
	sid := int32(4095)
	generator := NewIDGenerator(sid)
	time.Sleep(time.Second)
	for i := 0; i < b.N; i++ {
		_ = generator.NewID()
	}
}

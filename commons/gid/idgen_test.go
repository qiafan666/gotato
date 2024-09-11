package gid

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestGenerateID(t *testing.T) {

	sid := int32(4095)
	generator := NewID64Generator(sid)
	id := generator.NewID64()
	sid, createTime := ParseID64(id)
	fmt.Println(id)              // 7398101349412372480
	fmt.Println(sid, createTime) // 1 158992640
	fmt.Println(IsServerID64Valid(sid))
	serverID, at := ParseID64(id)
	fmt.Println(serverID, at) // 1 158992640

	sid1 := int32(0)
	idGenerator := NewID64Generator(sid1)
	id1 := idGenerator.NewID64()
	fmt.Println(id1) // 7398101349412372480
	fmt.Println(sid1, createTime)
	fmt.Println(IsServerID64Valid(sid))
}

func BenchmarkGenerateID(b *testing.B) {
	sid := int32(4095)
	generator := NewID64Generator(sid)
	time.Sleep(time.Second)
	for i := 0; i < b.N; i++ {
		_ = generator.NewID64()
	}
}

func TestRandID(t *testing.T) {

	fmt.Println(rand.Intn(65536))

	randID := RandID32()
	fmt.Println(randID)

	id, at := ParseID32(randID)
	fmt.Println(id, at)
}

package grpc

import (
	"math"
	"sync"
)

var (
	SeqMutex sync.Mutex
	Seq      uint32 = 0
)

func NewSeq() uint32 {
	SeqMutex.Lock()
	defer SeqMutex.Unlock()

	if Seq == math.MaxUint32 {
		Seq = 0
	}
	Seq++
	return Seq
}

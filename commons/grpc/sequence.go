package grpc

import (
	"math"
	"sync"
)

var (
	sequenceMutex sync.Mutex
	sequence      uint32 = 0
)

func NewSequence() uint32 {
	sequenceMutex.Lock()
	defer sequenceMutex.Unlock()

	if sequence == math.MaxUint32 {
		sequence = 0
	}
	sequence++
	return sequence
}

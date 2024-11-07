package gid

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestSnowFlakeUUID(t *testing.T) {
	worker, err := NewIDWorker(1)
	if err != nil {
		fmt.Println(err)
		return
	}

	id, err := worker.NextID()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("ID:", id)

	fmt.Println("Max ID:", rand.Intn(1<<(ServerIdBits)-1))
}

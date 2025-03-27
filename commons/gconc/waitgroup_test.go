package gconc_test

import (
	"fmt"
	"github.com/qiafan666/gotato/commons/gconc"
	"sync/atomic"
	"testing"
)

func TestExampleWaitGroup(t *testing.T) {
	var count atomic.Int64

	wg := gconc.NewWaitGroup()
	for i := 0; i < 10; i++ {
		wg.Go(func() {
			count.Add(1)
		})
	}
	wg.Wait()

	fmt.Println(count.Load())
	// Output:
	// 10
}

func TestExampleWaitGroup_WaitAndRecover(t *testing.T) {
	wg := gconc.NewWaitGroup()

	wg.Go(func() {
		panic("super bad thing")
	})

	recoveredPanic := wg.WaitAndRecover()
	fmt.Println(recoveredPanic.Value)
	// Output:
	// super bad thing
}

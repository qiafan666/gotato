package stream_test

import (
	"fmt"
	"github.com/qiafan666/gotato/commons/gconc/stream"
	"time"
)

func ExampleStream() {
	times := []int{20, 52, 16, 45, 4, 80}

	s := stream.New()
	for _, millis := range times {
		dur := time.Duration(millis) * time.Millisecond
		s.Go(func() stream.Callback {
			time.Sleep(dur)
			return func() { fmt.Println(dur) }
		})
	}
	s.Wait()

	// Output:
	// 20ms
	// 52ms
	// 16ms
	// 45ms
	// 4ms
	// 80ms
}

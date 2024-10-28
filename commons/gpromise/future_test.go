package gpromise_test

import (
	"fmt"
	"github.com/qiafan666/gotato/commons/gpromise"
	"github.com/qiafan666/gotato/commons/gtime"
	"log"
	"testing"
	"time"
)

func TestCommonFutureAfter(t *testing.T) {
	now := time.Now()
	pm := gpromise.NewManager(1, func() int { return 100 })

	p := pm.NewPromise("testPromise", func(context *gpromise.Context) {
		if context.Err != nil {
			log.Println("Promise testPromise error:", context.Err.Error())
		}
	})

	safeFinish := func(res []int) {
		var total int
		for _, re := range res {
			total += re
		}
		fmt.Println("Promise testPromise result:", total)
	}

	future := gpromise.NewCommonFuture("testFuture")

	future.OnDo = func() error {
		log.Println("Future testFuture do")
		time.Sleep(2 * time.Second) // 模拟耗时任务
		pm.Process(gpromise.GetPfId(p.Id, future.Id()), []interface{}{1, 2, 3}, nil)
		return nil
	}

	future.OnCallBack = func(result []interface{}) error {
		log.Println("Future testFuture callback:", result)
		resultInt := make([]int, len(result))
		for i, re := range result {
			resultInt[i] = re.(int)
		}
		safeFinish(resultInt)
		return nil
	}

	go func() {
		p.Push(future)
		p.Start() // 异步启动Promise
	}()

	time.Sleep(3 * time.Second) // 等待异步执行完成
	fmt.Println(gtime.Since(now))
}

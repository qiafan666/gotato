package gpromise_test

import (
	"fmt"
	"github.com/qiafan666/gotato/commons/gpromise"
	"github.com/qiafan666/gotato/commons/gtime/logictime"
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
		} else {
			log.Println("Promise testPromise success")
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
		pm.Process(future.GetPfId(), []interface{}{1, 2, 3}, nil)
		return nil
	}

	done := make(chan struct{})

	// 修改 future.OnCallBack，使其在回调结束后向 `done` 通道发送信号
	future.OnCallBack = func(result []interface{}) error {
		log.Println("Future testFuture callback:", result)
		resultInt := make([]int, len(result))
		for i, re := range result {
			resultInt[i] = re.(int)
		}
		safeFinish(resultInt)
		close(done) // 通知主线程操作完成
		return nil
	}

	// 异步启动Promise
	p.Push(future)
	p.Start()

	// 等待 Promise 完成，不再依赖 time.Sleep
	<-done
	fmt.Println(logictime.Since(now))
}

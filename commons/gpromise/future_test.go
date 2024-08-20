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

			if context.Err != nil {
				log.Println("Promise testPromise error:", context.Err.Error())
			}
		}
	})

	//实际业务逻辑
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
		futureOndoCallback(pm, gpromise.GetPfId(p.Id, future.Id()))
		return nil
	}

	fmt.Println("逻辑异步")
	future.OnCallBack = func(result []interface{}) error {
		log.Println("Future testFuture callback:", result)
		resultInt := make([]int, len(result))
		for i, re := range result {
			resultInt[i] = re.(int)
		}
		safeFinish(resultInt)
		return nil
	}

	p.Push(future)
	p.Start()

	fmt.Println(gtime.Since(now))
}

func futureOndoCallback(pm *gpromise.Manager, pfId uint64) {
	//回调
	pm.Process(pfId, []interface{}{1, 2, 3}, nil)
}

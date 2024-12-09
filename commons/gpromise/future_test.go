package gpromise_test

import (
	"fmt"
	"github.com/qiafan666/gotato/commons/gpromise"
	"github.com/qiafan666/gotato/commons/gtime/logictime"
	"log"
	"testing"
	"time"
)

type StdLogger struct{}

func (l *StdLogger) PromiseErrorF(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

func (l *StdLogger) PromiseWarnF(format string, args ...interface{}) {
	log.Printf("[WARN] "+format, args...)
}

func (l *StdLogger) PromiseInfoF(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

func (l *StdLogger) PromiseDebugF(format string, args ...interface{}) {
	log.Printf("[DEBUG] "+format, args...)
}

func TestCommonFutureAfter(t *testing.T) {
	now := time.Now()
	pm := gpromise.NewManager(1, func() int { return 100 }, &StdLogger{})

	p := pm.NewPromise("promise", func(context *gpromise.Context) {
		if context.Err != nil {
			log.Println("异步管理执行失败:", context.Err.Error())
		} else {
			log.Println("异步管理执行成功")
		}
	})

	safeFinish := func(res []int) {
		var total int
		for _, re := range res {
			total += re
		}
		fmt.Println("执行成功函数:", total)
	}

	future := gpromise.NewCommonFuture("future")

	future.OnDo = func() error {
		log.Println("future OnDo")
		time.Sleep(2 * time.Second) // 模拟耗时任务
		pm.Process(future.GetPfId(), []interface{}{1, 2, 3}, nil)
		return nil
	}

	done := make(chan struct{})

	// 修改 future.OnCallBack，使其在回调结束后向 `done` 通道发送信号
	future.OnCallBack = func(result []interface{}) error {
		log.Println("future OnCallBack:", result)
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

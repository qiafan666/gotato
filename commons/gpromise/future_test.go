package gpromise_test

import (
	"context"
	"fmt"
	"github.com/qiafan666/gotato/commons/glog"
	"github.com/qiafan666/gotato/commons/gpromise"
	"github.com/qiafan666/gotato/commons/gtime/logictime"
	"go.uber.org/zap"
	"log"
	"testing"
	"time"
)

type logger struct{}

func (l *logger) ErrorF(ctx context.Context, format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[ERROR] [%s] ", l.Prefix())+glog.GetTraceId(ctx)+format, args...)
	} else {
		l.Logger().Errorf(fmt.Sprintf(l.Prefix())+glog.GetTraceId(ctx)+format, args...)
	}
}
func (l *logger) WarnF(ctx context.Context, format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[WARN] [%s] ", l.Prefix())+glog.GetTraceId(ctx)+format, args...)
	} else {
		l.Logger().Warnf(fmt.Sprintf(l.Prefix())+glog.GetTraceId(ctx)+format, args...)
	}
}
func (l *logger) InfoF(ctx context.Context, format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[INFO] [%s] ", l.Prefix())+glog.GetTraceId(ctx)+format, args...)
	} else {
		l.Logger().Infof(fmt.Sprintf(l.Prefix())+glog.GetTraceId(ctx)+format, args...)
	}
}
func (l *logger) DebugF(ctx context.Context, format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[DEBUG] [%s] ", l.Prefix())+glog.GetTraceId(ctx)+format, args...)
	} else {
		l.Logger().Debugf(fmt.Sprintf(l.Prefix())+glog.GetTraceId(ctx)+format, args...)
	}
}
func (l *logger) Logger() *zap.SugaredLogger {
	return nil
}
func (l *logger) Prefix() string {
	return "promise"
}

func TestCommonFutureAfter(t *testing.T) {
	now := time.Now()
	pm := gpromise.NewManager(1, func() int { return 100 }, &logger{})

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
		return nil
	}

	afterFuture := gpromise.NewCommonFuture("afterFuture")
	afterFuture.OnDo = func() error {
		log.Println("afterFuture OnDo")
		time.Sleep(1 * time.Second) // 模拟耗时任务
		pm.Process(afterFuture.GetPfId(), []interface{}{4, 5, 6}, nil)
		return nil
	}

	// 修改 afterFuture.OnCallBack，使其在回调结束后向 `done` 通道发送信号
	afterFuture.OnCallBack = func(result []interface{}) error {
		log.Println("afterFuture OnCallBack:", result)
		resultInt := make([]int, len(result))
		for i, re := range result {
			resultInt[i] = re.(int)
		}
		safeFinish(resultInt)
		close(done)
		return nil
	}

	// 异步启动Promise
	p.Push(future)
	future.After(afterFuture)

	p.Start()

	// 等待 Promise 完成，不再依赖 time.Sleep
	<-done
	fmt.Println(logictime.Since(now))
}

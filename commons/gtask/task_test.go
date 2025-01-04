package gtask

import (
	"context"
	"fmt"
	"github.com/qiafan666/gotato/commons/gcommon"
	"go.uber.org/zap"
	"log"
	"strconv"
	"testing"
	"time"
)

type logger struct{}

func (l *logger) ErrorF(ctx context.Context, format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[ERROR] [%s] ", l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	} else {
		l.Logger().Errorf(fmt.Sprintf(l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	}
}
func (l *logger) WarnF(ctx context.Context, format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[WARN] [%s] ", l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	} else {
		l.Logger().Warnf(fmt.Sprintf(l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	}
}
func (l *logger) InfoF(ctx context.Context, format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[INFO] [%s] ", l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	} else {
		l.Logger().Infof(fmt.Sprintf(l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	}
}
func (l *logger) DebugF(ctx context.Context, format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[DEBUG] [%s] ", l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	} else {
		l.Logger().Debugf(fmt.Sprintf(l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	}
}
func (l *logger) Logger() *zap.SugaredLogger {
	return nil
}
func (l *logger) Prefix() string {
	return "task"
}

func TestTask(t *testing.T) {

	// 初始化任务池
	taskNum := 5
	chanNum := 50
	InitDefaultPool(taskNum, chanNum, &logger{})

	// 添加任务
	for i := 0; i < 10; i++ {
		id := i
		err := AddTask(func() {
			fmt.Printf("Task %d executed\n", id)
		}, func() {
			fmt.Printf("Task %d callback executed\n", id)
		}, strconv.Itoa(id))
		if err != nil {
			t.Fatalf("Failed to add task: %v", err)
		}
	}

	time.Sleep(time.Second * 1)
	// 停止任务池
	StopDefaultPool()
}

package gtask

import (
	"fmt"
	"go.uber.org/zap"
	"log"
	"strconv"
	"testing"
	"time"
)

type logger struct{}

func (l *logger) ErrorF(format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[ERROR] [%s] ", l.Prefix())+format, args...)
	} else {
		l.Logger().Errorf(fmt.Sprintf("[%s] ", l.Prefix())+format, args...)
	}
}
func (l *logger) WarnF(format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[WARN] [%s] ", l.Prefix())+format, args...)
	} else {
		l.Logger().Warnf(fmt.Sprintf("[%s] ", l.Prefix())+format, args...)
	}
}
func (l *logger) InfoF(format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[INFO] [%s] ", l.Prefix())+format, args...)
	} else {
		l.Logger().Infof(fmt.Sprintf("[%s] ", l.Prefix())+format, args...)
	}
}
func (l *logger) DebugF(format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[DEBUG] [%s] ", l.Prefix())+format, args...)
	} else {
		l.Logger().Debugf(fmt.Sprintf("[%s] ", l.Prefix())+format, args...)
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

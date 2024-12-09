package gtask

import (
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"
)

type StdLogger struct{}

func (l *StdLogger) TaskErrorF(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

func (l *StdLogger) TaskWarnF(format string, args ...interface{}) {
	log.Printf("[WARN] "+format, args...)
}

func TestTask(t *testing.T) {

	// 初始化任务池
	taskNum := 5
	chanNum := 50
	logger := &StdLogger{}
	InitDefaultPool(taskNum, chanNum, logger)

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

package gtask

import (
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"
)

type StdLogger struct{}

func (l *StdLogger) ErrorF(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}
func (l *StdLogger) WarnF(format string, args ...interface{}) {
	log.Printf("[WARN] "+format, args...)
}
func (l *StdLogger) InfoF(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}
func (l *StdLogger) DebugF(format string, args ...interface{}) {
	log.Printf("[DEBUG] "+format, args...)
}

func TestTask(t *testing.T) {

	// 初始化任务池
	taskNum := 5
	chanNum := 50
	InitDefaultPool(taskNum, chanNum, &StdLogger{})

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

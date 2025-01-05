package gtask

import (
	"fmt"
	"github.com/qiafan666/gotato/commons/gface"
	"strconv"
	"testing"
	"time"
)

func TestTask(t *testing.T) {

	// 初始化任务池
	taskNum := 5
	chanNum := 50
	InitDefaultPool(taskNum, chanNum, gface.NewLogger("task", nil))

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

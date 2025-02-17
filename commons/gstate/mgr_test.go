package gstate

import (
	"fmt"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/gface"
	"testing"
)

func TestStateMachine(t *testing.T) {

	// 创建状态机管理器
	stateMachineManager := NewStateMachineManager(gface.NewLogger("state_machine", nil))

	// 创建任务状态机
	taskMachine := NewStateMachine(gcommon.SetRequestId("create_event"))
	// 添加状态转换
	taskMachine.AddTransition("TaskCreated", "TaskInProgress", "Start", func(data []interface{}) {
		m := data[0].(map[string]interface{})
		fmt.Println("任务开始执行", m["1"], m["2"], m["3"])
	})
	taskMachine.AddTransition("TaskInProgress", "TaskCompleted", "Complete", func(data []interface{}) {
		m := data[0].(map[string]interface{})
		fmt.Println("任务完成", m["1"], m["2"], m["3"])
	})
	taskMachine.AddTransition("TaskInProgress", "TaskCancelled", "Cancel", func(data []interface{}) {
		m := data[0].(map[string]interface{})
		fmt.Println("任务取消", m["1"], m["2"], m["3"])
	})

	stateMachineManager.AddMachine("CreateTask", taskMachine)

	data := make(map[string]interface{})
	data["1"] = 1
	data["2"] = 2
	data["3"] = 3
	datas := []interface{}{data}
	// 测试任务状态流程
	// 任务开始
	stateMachineManager.HandleEvent("CreateTask", "Start", datas) // 输出：状态从 TaskCreated 转换到 TaskInProgress
	// 任务完成
	stateMachineManager.HandleEvent("CreateTask", "Complete", datas) // 输出：状态从 TaskInProgress 转换到 TaskCompleted
	// 获取当前状态
	fmt.Println("当前状态：", taskMachine.GetState()) // 输出：TaskCompleted

	// 测试任务取消
	cancelMachine := NewStateMachine(gcommon.SetRequestId("cancel_event"))
	cancelMachine.AddTransition("TaskCreated", "TaskInProgress", "Start", func(data []interface{}) {
		fmt.Println("任务开始执行")
	})
	cancelMachine.AddTransition("TaskInProgress", "TaskCancelled", "Cancel", func(data []interface{}) {
		fmt.Println("任务已取消")
	})
	stateMachineManager.AddMachine("Cancel", cancelMachine)

	// 任务开始
	stateMachineManager.HandleEvent("Cancel", "Start", nil) // 输出：状态从 TaskCreated 转换到 TaskInProgress
	// 任务取消
	stateMachineManager.HandleEvent("Cancel", "Cancel", nil) // 输出：状态从 TaskInProgress 转换到 TaskCancelled
	// 获取当前状态
	fmt.Println("当前状态：", cancelMachine.GetState()) // 输出：TaskCancelled

	// 测试无效事件
	invalidMachine := NewStateMachine(gcommon.SetRequestId("invalid_event"))
	stateMachineManager.AddMachine("Invalid", invalidMachine)

	invalidMachine.AddTransition("TaskCreated", "TaskInProgress", "Start", func(data []interface{}) {
		fmt.Println("任务开始执行")
	})

	stateMachineManager.HandleEvent("Invalid", "Complete", nil) // 输出：事件 Complete 在状态机中不存在

	stateMachineManager.RemoveMachine("CreateTask")
	stateMachineManager.RemoveMachine("Cancel")
	stateMachineManager.RemoveMachine("Invalid")
	machine, b := stateMachineManager.GetMachine("CreateTask")
	if b || machine != nil {
		t.Error("获取到任务状态机")
	} else {
		t.Log("未获取到任务状态机")
	}
}

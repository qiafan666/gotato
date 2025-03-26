package gstate

import (
	"context"
	"github.com/qiafan666/gotato/commons/gface"
)

// State 表示状态机中的状态
type State string

// Event 表示事件
type Event string

// Transition 定义一个结构体来记录状态转换信息
type Transition struct {
	FromState State
	ToState   State
}

// StateMachine 状态机结构体
type StateMachine struct {
	state       State
	transitions map[State]map[Event]*Transition
	handlers    map[Event]func(data []interface{}) // 事件处理函数
	logger      gface.ILogger                      // 日志记录器
	ctx         context.Context                    // 上下文
}

// NewStateMachine 创建一个新的状态机
func NewStateMachine(ctx context.Context) *StateMachine {
	return &StateMachine{
		transitions: make(map[State]map[Event]*Transition),
		handlers:    make(map[Event]func(data []interface{})),
		ctx:         ctx,
	}
}

// AddTransition 添加状态转换
func (sm *StateMachine) AddTransition(fromState State, toState State, event Event, handler func(data []interface{})) {

	if fromState == "" || toState == "" || event == "" {
		sm.logger.ErrorF(sm.ctx, "fromState or toState or event is empty,fromState:%v,toState:%v,event:%v", fromState, toState, event)
		return
	}

	// 初始状态为空，则设置初始状态,所以AddTransition必须从初始状态开始
	if sm.state == "" {
		sm.state = fromState
	}

	if sm.transitions[fromState] == nil {
		sm.transitions[fromState] = make(map[Event]*Transition)
	}
	sm.transitions[fromState][event] = &Transition{
		FromState: fromState,
		ToState:   toState,
	}

	// 事件处理函数
	sm.handlers[event] = handler
}

// HandleEvent 处理事件并转换状态
func (sm *StateMachine) handleEvent(name string, event Event, data []interface{}) {

	// 查找当前状态下，是否有对应的事件转换
	if memoryState, ok := sm.transitions[sm.state][event]; ok {
		sm.logger.InfoF(sm.ctx, "name：%s，event：%v，current state: %v，event state switch: %v -> %v", name, event, sm.state, memoryState.FromState, memoryState.ToState)

		if sm.state == memoryState.FromState {
			// 状态转换
			if handler, exists := sm.handlers[event]; exists && handler != nil {
				handler(data) // 执行事件处理函数
				sm.state = memoryState.ToState
			} else {
				sm.logger.WarnF(sm.ctx, "name：%s，event：%v，current state: %v，event state switch: %v -> %v，but no handler found", name, event, sm.state, memoryState.FromState, memoryState.ToState)
			}
		} else {
			sm.logger.WarnF(sm.ctx, "name：%s，event：%v，current state: %v，event state switch: %v -> %v，but current state is not equal to event state switch from state", name, event, sm.state, memoryState.FromState, memoryState.ToState)
		}
	} else {
		sm.logger.ErrorF(sm.ctx, "name：%s，event:%v，current state: %v，no transition found", name, event, sm.state)
	}
}

// GetState 打印当前状态
func (sm *StateMachine) GetState() State {
	return sm.state
}

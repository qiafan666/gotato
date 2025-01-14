package gstate

import (
	"context"
	"github.com/qiafan666/gotato/commons/gface"
)

// State 表示状态机中的状态
type State string

// Event 表示事件
type Event string

// StateMachine 状态机结构体
type StateMachine struct {
	state       State
	transitions map[State]map[Event]State
	handlers    map[Event]func(data []interface{}) // 事件处理函数
	logger      gface.Logger                       // 日志记录器
	ctx         context.Context                    // 上下文
}

// NewStateMachine 创建一个新的状态机
func NewStateMachine(ctx context.Context, initialState State) *StateMachine {
	return &StateMachine{
		state:       initialState,
		transitions: make(map[State]map[Event]State),
		handlers:    make(map[Event]func(data []interface{})),
		ctx:         ctx,
	}
}

// AddTransition 添加状态转换
func (sm *StateMachine) AddTransition(fromState State, event Event, toState State) {
	if sm.transitions[fromState] == nil {
		sm.transitions[fromState] = make(map[Event]State)
	}
	sm.transitions[fromState][event] = toState
}

// SetHandler 设置事件处理函数
func (sm *StateMachine) SetHandler(event Event, handler func(data []interface{})) {
	sm.handlers[event] = handler
}

// HandleEvent 处理事件并转换状态
func (sm *StateMachine) handleEvent(name string, event Event, data []interface{}) {

	// 查找当前状态下，是否有对应的事件转换
	if nextState, ok := sm.transitions[sm.state][event]; ok {

		sm.logger.InfoF(sm.ctx, "name：%s，event：%v，state switch: %v -> %v", name, event, sm.state, nextState)
		sm.state = nextState
		if handler, exists := sm.handlers[event]; exists {
			handler(data) // 执行事件处理函数
		}
	} else {
		sm.logger.ErrorF(sm.ctx, "name：%s，event:%v，current state: %v，no transition found", name, event, sm.state)
	}
}

// GetState 打印当前状态
func (sm *StateMachine) GetState() State {
	return sm.state
}

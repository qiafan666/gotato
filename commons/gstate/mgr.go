package gstate

import (
	"github.com/qiafan666/gotato/commons/gface"
	"sync"
)

// StateMachineManager 状态机管理器
type StateMachineManager struct {
	machines map[string]*StateMachine
	mu       sync.Mutex
	logger   gface.Logger
}

// NewStateMachineManager 创建一个新的状态机管理器
func NewStateMachineManager(logger gface.Logger) *StateMachineManager {
	return &StateMachineManager{
		machines: make(map[string]*StateMachine),
		logger:   logger,
	}
}

// AddMachine 向管理器中添加一个状态机
func (smm *StateMachineManager) AddMachine(name string, sm *StateMachine) {
	smm.mu.Lock()
	defer smm.mu.Unlock()
	sm.logger = smm.logger
	if _, exists := smm.machines[name]; exists {
		smm.logger.ErrorF(nil, "state machine %s already exists", name)
		return
	}
	smm.machines[name] = sm
}

// RemoveMachine 从管理器中移除一个状态机
func (smm *StateMachineManager) RemoveMachine(name string) {
	smm.mu.Lock()
	defer smm.mu.Unlock()
	delete(smm.machines, name)
}

// GetMachine 获取指定名称的状态机实例
func (smm *StateMachineManager) GetMachine(name string) (*StateMachine, bool) {
	smm.mu.Lock()
	defer smm.mu.Unlock()
	sm, exists := smm.machines[name]
	return sm, exists
}

// HandleEvent 处理某个特定状态机实例的事件
func (smm *StateMachineManager) HandleEvent(name string, event Event, data []interface{}) {
	smm.mu.Lock()
	defer smm.mu.Unlock()

	if sm, exists := smm.machines[name]; exists {
		sm.handleEvent(name, event, data)
	} else {
		smm.logger.ErrorF(nil, "state machine %s not found", name)
	}
}

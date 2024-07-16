package utils

import (
	"errors"
	"fmt"
	"github.com/asaskevich/EventBus"
	"reflect"
)

type MsgBus struct {
	topic       string       // 主题
	bus         EventBus.Bus // 事件总线
	sync        bool         // 同步
	transaction bool         // 事务
	once        bool         // 一次性
	callback    reflect.Value
}

func NewMsgBus(topic string, sync bool, transaction bool, once bool) *MsgBus {
	return &MsgBus{
		topic:       topic,
		bus:         EventBus.New(),
		sync:        sync,
		transaction: transaction,
		once:        once,
	}
}

// Publish 发布
func (m *MsgBus) Publish(args ...interface{}) error {
	if m.callback.IsValid() {
		if err := validateArgs(m.callback, args...); err != nil {
			return err
		}
	}
	m.bus.Publish(m.topic, args...)
	return nil
}

// Subscribe 订阅
func (m *MsgBus) Subscribe(f interface{}) error {

	callback := reflect.ValueOf(f)
	if callback.Kind() != reflect.Func {
		return errors.New("provided subscription is not a function")
	}
	m.callback = callback

	if m.once {
		if m.sync {
			return m.bus.SubscribeOnce(m.topic, f)
		} else {
			return m.bus.SubscribeOnceAsync(m.topic, f)
		}
	} else {
		if m.sync {
			return m.bus.Subscribe(m.topic, f)
		} else {
			return m.bus.SubscribeAsync(m.topic, f, m.transaction)
		}
	}
}

// Unsubscribe 取消订阅
func (m *MsgBus) Unsubscribe(f interface{}) error {
	return m.bus.Unsubscribe(m.topic, f)
}

// WaitAsync 等待异步执行完成
func (m *MsgBus) WaitAsync() {
	m.bus.WaitAsync()
}

func validateArgs(callback reflect.Value, args ...interface{}) error {
	funcType := callback.Type()
	if len(args) != funcType.NumIn() {
		return fmt.Errorf("argument count mismatch: expected %d, got %d", funcType.NumIn(), len(args))
	}
	for i, arg := range args {
		if arg == nil {
			continue
		}
		if reflect.TypeOf(arg) != funcType.In(i) {
			return fmt.Errorf("argument type mismatch at index %d: expected %s, got %s", i, funcType.In(i), reflect.TypeOf(arg))
		}
	}
	return nil
}

package gcommon

import (
	"fmt"
	"runtime"
)

// PrintPanicStack recover并打印堆栈
// 用法: defer utils.PrintPanicStack()，注意defer func() {util.PrintPanicStack()}是无效的!
func PrintPanicStack() string {
	if r := recover(); r != nil {
		stack := Stack()
		return fmt.Sprintf("%v: %s", r, stack)
	}
	return ""
}

// Stack 返回调用堆栈
func Stack() string {
	buf := make([]byte, 2048)
	l := runtime.Stack(buf, false)
	return BytesToStr(buf[:l])
}

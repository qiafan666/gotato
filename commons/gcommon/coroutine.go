package gcommon

import (
	"fmt"
	"log"
	"reflect"
	"runtime/debug"
)

// f - 协程函数，用于处理panic，recoverNum为-1表示一直恢复，recoverNum为0表示不恢复
func f(callback interface{}, recoverNum int, args ...interface{}) {
	defer func() {
		if r := recover(); r != nil {
			var coreInfo string
			coreInfo = string(debug.Stack())
			coreInfo += "\n" + fmt.Sprintf("Core information is %v\n", r)
			log.Println(coreInfo)
			if recoverNum == -1 || recoverNum-1 >= 0 {
				recoverNum -= 1
				go f(callback, recoverNum, args...)
			}
		}
	}()

	v := reflect.ValueOf(callback)
	if v.Kind() != reflect.Func {
		panic("not a function")
	}
	vargs := make([]reflect.Value, len(args))
	for i, arg := range args {
		vargs[i] = reflect.ValueOf(arg)
	}

	v.Call(vargs)
}

// Go - 启动一个协程
func Go(callback interface{}, args ...interface{}) {
	go f(callback, 0, args...)
}

// GoRecover -1表示一直恢复
func GoRecover(callback interface{}, recoverNum int, args ...interface{}) {
	go f(callback, recoverNum, args...)
}

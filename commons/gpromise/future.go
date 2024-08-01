package gpromise

import (
	"container/list"
	"fmt"
	"log"
)

type IFuture interface {
	Do() error
	CallBack(args []interface{}) error

	Id() uint32
	SetStatus(status int)
	GetStatus() int
	GetPfId() uint64
	GetPromiseContext() *Context
	PushAfter(future IFuture)

	init(promise *Promise)
	isFinish() bool
	isCombine() bool

	Name() string
}

type Future struct {
	id           uint32
	status       int
	promise      *Promise
	fatherFuture IFuture
	name         string
}

func NewFuture(name string) *Future {
	return &Future{name: name}
}

func (f *Future) Name() string {
	return f.name
}

func (f *Future) init(promise *Promise) {
	if f == nil {
		log.Println("Error: future is nil")
		return
	}
	if promise == nil {
		log.Println("Error: promise is nil")
		return
	}
	f.id = promise.futureIdIndex
	f.promise = promise
	f.status = FutureStatusNull

	promise.futureIdIndex++
}

func (f *Future) Id() uint32 {
	return f.id
}

func (f *Future) SetStatus(status int) {
	f.status = status
}

func (f *Future) GetStatus() int {
	return f.status
}

func (f *Future) GetPfId() uint64 {
	promiseId := f.promise.Id
	futureId := f.id
	pfId := uint64(promiseId) << 32
	pfId += uint64(futureId)
	return pfId
}

func (f *Future) GetPromiseContext() *Context {
	return f.promise.context
}

func (f *Future) PushAfter(future IFuture) {
	var e *list.Element
	var ok bool
	e, ok = f.promise.futureMap[f.id]
	if !ok {
		log.Println(fmt.Sprintf("Error: not find future[%v] in promise[%v]", f.id, f.promise.Id))
		return
	} else {
		future.init(f.promise)
		e = f.promise.futures.InsertAfter(future, e)
		f.promise.futureMap[future.Id()] = e
	}
}

func (f *Future) isFinish() bool {
	return f.status == FutureStatusFinish
}

func (f *Future) isCombine() bool {
	return false
}

type CommonFuture struct {
	*Future
	OnDo       func() error
	OnCallBack func(args []interface{}) error
}

func NewCommonFuture(name string) *CommonFuture {
	return &CommonFuture{
		Future: &Future{name: name},
	}
}

func (f *CommonFuture) Do() error {
	if f.OnDo != nil {
		log.Println(fmt.Sprintf("Debug: do future name:%v", f.name))
		return f.OnDo()
	}

	return nil
}

func (f *CommonFuture) CallBack(args []interface{}) error {
	if f.OnCallBack != nil {
		log.Println(fmt.Sprintf("Debug: callback future name:%v", f.name))
		return f.OnCallBack(args)
	}

	return nil
}

// After 不要在Future的CallBack 使用Future.After
func (f *CommonFuture) After(af *CommonFuture) *CommonFuture {
	f.wrapCallback(func() {
		f.PushAfter(af)
	})
	return af
}

func (f *CommonFuture) AfterIf(condFn func() bool, af, ef *CommonFuture) {
	f.wrapCallback(func() {
		if condFn() {
			f.PushAfter(af)
		} else {
			f.PushAfter(ef)
		}
	})
}

func (f *CommonFuture) wrapCallback(wrapper func()) {
	prev := f.OnCallBack
	f.OnCallBack = func(args []interface{}) error {
		er := prev(args)
		if er != nil {
			return er

		}
		wrapper()
		return nil
	}
}

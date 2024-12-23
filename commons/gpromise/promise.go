package gpromise

import (
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"github.com/qiafan666/gotato/commons/gface"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	FutureStatusNull   = 0 //未开始
	FutureStatusDo     = 1 //执行中
	FutureStatusFinish = 2 //结束
)

var ContinueErr = errors.New("CONTINUE")

const MaxRuntime = 30 // promise默认允许最大执行时间

type StatResult struct {
	Name     string
	Duration time.Duration
}

type Context struct {
	Args  []interface{}
	Err   error
	Kw    map[string]interface{}
	trace strings.Builder
}

func (pc *Context) Get(k string) interface{} {
	return pc.Kw[k]
}

func (pc *Context) Save(k string, v interface{}) {
	if pc.Kw == nil {
		pc.Kw = make(map[string]interface{})
	}
	pc.Kw[k] = v
}

func (pc *Context) Trace(name string, id uint32) {
	if name == "" {
		return
	}
	if id == 0 {
		pc.trace.WriteString(name)
	} else {
		pc.trace.WriteString(fmt.Sprintf("->%v(%v)", name, id))
	}
}

func (pc *Context) GetTrace() string {
	return pc.trace.String()
}

type Promise struct {
	Id            uint32
	Name          string
	futures       *list.List
	futureMap     map[uint32]*list.Element
	pm            *Manager
	createTime    time.Time
	cbFunc        func(context *Context)
	context       *Context
	futureIdIndex uint32
	logger        gface.Logger
}

type Callback struct {
	PfId  uint64
	Owner int64
	Args  []interface{}
}

type Manager struct {
	owner          int64
	promises       map[uint32]*Promise
	promiseIdIndex uint32

	statResults          []*StatResult
	maxResult            *StatResult
	lastPrintPromiseStat time.Time
	warnThresholdFunc    func() int // 获取需要打印警告的promise耗时时间

	onAllPromiseOver func()

	logger gface.Logger
}

func NewManager(owner int64, f func() int, logger gface.Logger) *Manager {
	pm := &Manager{
		owner:                owner,
		promises:             make(map[uint32]*Promise),
		promiseIdIndex:       1,
		statResults:          []*StatResult{},
		lastPrintPromiseStat: time.Now(),
		warnThresholdFunc:    f,
		logger:               logger,
	}

	return pm
}

func (pm *Manager) NewPromise(name string, cbFunc func(context *Context), args ...interface{}) *Promise {
	p := &Promise{
		Id:         pm.promiseIdIndex,
		Name:       name,
		futures:    list.New(),
		futureMap:  make(map[uint32]*list.Element),
		pm:         pm,
		createTime: time.Now(),
		context: &Context{
			Args:  args,
			trace: strings.Builder{},
		},
		futureIdIndex: 1,
		logger:        pm.logger,
	}
	p.context.Trace(name, 0)
	pm.promiseIdIndex++
	pm.promises[p.Id] = p

	p.cbFunc = func(context *Context) {
		if cbFunc != nil {
			cbFunc(p.context)
		}

		statResult := &StatResult{
			Name:     name,
			Duration: time.Since(p.createTime),
		}

		pm.statResults = append(pm.statResults, statResult)

		if pm.maxResult == nil || statResult.Duration > pm.maxResult.Duration {
			pm.maxResult = statResult
		}
	}

	return p
}

func (pm *Manager) SetAllPromiseOverFunc(onOver func()) {
	if onOver == nil {
		return
	}

	if len(pm.promises) == 0 {
		onOver()
		return
	}

	pm.onAllPromiseOver = onOver
}

func (pm *Manager) DeletePromise(Id uint32) {
	delete(pm.promises, Id)
	if len(pm.promises) == 0 && pm.onAllPromiseOver != nil {
		pm.onAllPromiseOver()
	}
}

func (pm *Manager) GetPromise(Id uint32) *Promise {
	return pm.promises[Id]
}

func (p *Promise) Push(future IFuture) {
	if future == nil {
		p.logger.ErrorF("Promise: future is nil")
		return
	}

	if p == nil {
		p.logger.ErrorF("Promise: promise is nil")
		return
	}

	future.init(p)
	e := p.futures.PushBack(future)
	p.futureMap[future.Id()] = e
}

func (p *Promise) GetAllPfId() string {
	var sPfId bytes.Buffer
	promiseId := p.Id
	for k := range p.futureMap {
		futureId := k
		pfId := uint64(promiseId) << 32
		pfId += uint64(futureId)
		sPfId.WriteString(strconv.FormatUint(pfId, 10))
		sPfId.WriteString(" ")
	}
	return sPfId.String()
}

func (p *Promise) Start() {
	promiseFinish := false
	e := p.futures.Front()
	if e == nil {
		promiseFinish = true
	}

	for !promiseFinish {
		if _, exists := p.pm.promises[p.Id]; !exists {
			break
		}
		f, ok := e.Value.(IFuture)
		if !ok {
			p.context.Err = errors.New("element is not future")
			promiseFinish = true
			break
		}

		f.SetStatus(FutureStatusDo)

		do := func() (errInfo error) {
			defer func() {
				if r := recover(); r != nil {
					errInfo = errors.New(fmt.Sprintf("%v:%v", r, GetStack()))
				}
			}()
			f.GetPromiseContext().Trace(f.Name(), f.Id())
			return f.Do()
		}

		er := do()
		if er != nil {
			if !errors.Is(er, ContinueErr) {
				p.context.Err = er
				p.logger.WarnF("Promise: Owner[%v] promise[%v] future[%v] %v do fail[%v]",
					p.pm.owner, p.Id, f.Id(), GetPfId(p.Id, f.Id()), p.context.Err.Error())
				promiseFinish = true
				break
			} else {
				f.SetStatus(FutureStatusFinish)
			}
		}

		if f.isFinish() {
			e = e.Next()
			if e == nil {
				promiseFinish = true
			} else {
				continue
			}
		}

		break
	}

	if promiseFinish {
		// 检查是否仍然存在于 Manager 中
		if _, exists := p.pm.promises[p.Id]; exists {
			if p.cbFunc != nil {
				p.cbFunc(p.context)
			}
			p.pm.DeletePromise(p.Id)
		}
	}
}

func (p *Promise) Error() error {
	return p.context.Err
}

func (p *Promise) Trace() string {
	return p.context.trace.String()
}

func (pm *Manager) Process(pfId uint64, args []interface{}, errInfo error) {
	promiseId, futureId := GetPromiseFutureId(pfId)

	p, ok := pm.promises[promiseId]
	if !ok {
		p.logger.WarnF("Manager: Owner[%v] promise[%v] future[%v] %v is not exist",
			pm.owner, promiseId, futureId, pfId)
		return
	}

	e, ok := p.futureMap[futureId]
	if !ok {
		p.logger.ErrorF("Manager: Owner[%v] promise[%v] future[%v] %v is not exist",
			pm.owner, promiseId, futureId, pfId)
		return
	}

	f, ok := e.Value.(IFuture)
	if !ok {
		p.logger.ErrorF("Manager: Owner[%v] promise[%v] future[%v] %v element is not future",
			pm.owner, promiseId, futureId, pfId)
		p.context.Err = errors.New("element is not future")
		if p.cbFunc != nil {
			p.cbFunc(p.context)
		}
		pm.DeletePromise(p.Id)
		return
	}

	if errInfo != nil {
		p.logger.ErrorF("Manager: Owner[%v] promise[%v] future[%v] %v do fail[%v]",
			pm.owner, promiseId, futureId, pfId, errInfo.Error())
		p.context.Err = errInfo
		if p.cbFunc != nil {
			p.cbFunc(p.context)
		}
		pm.DeletePromise(p.Id)
		return
	}

	cb := func(args []interface{}) (errInfo error) {
		defer func() {
			if r := recover(); r != nil {
				errInfo = errors.New(fmt.Sprintf("%v:%v", r, GetStack()))
			}
		}()
		return f.CallBack(args)
	}
	er := cb(args)
	if er != nil {
		if !errors.Is(er, ContinueErr) {
			p.context.Err = er
			p.logger.WarnF("Manager: Owner[%v] promise[%v] future[%v] %v CallBack fail[%v]",
				pm.owner, promiseId, futureId, pfId, p.context.Err.Error())
			if p.cbFunc != nil {
				p.cbFunc(p.context)
			}
			pm.DeletePromise(p.Id)
			return
		}
	}
	f.SetStatus(FutureStatusFinish)

	promiseFinish := false
	for {
		e = e.Next()
		if e == nil {
			promiseFinish = true
			break
		}

		if _, exists := p.pm.promises[p.Id]; !exists {
			promiseFinish = true
			break
		}

		f, ok = e.Value.(IFuture)
		if !ok {
			p.context.Err = errors.New("element is not future")
			promiseFinish = true
			break
		}

		f.SetStatus(FutureStatusDo)

		do := func() (errInfo error) {
			defer func() {
				if r := recover(); r != nil {
					errInfo = errors.New(fmt.Sprintf("%v:%v", r, GetStack()))
				}
			}()
			f.GetPromiseContext().Trace(f.Name(), f.Id())
			return f.Do()
		}

		er = do()
		if er != nil {
			if !errors.Is(er, ContinueErr) {
				p.context.Err = er
				p.logger.WarnF("Manager: Owner[%v] promise[%v] future[%v] %v do fail[%v]",
					pm.owner, p.Id, f.Id(), pfId, p.context.Err.Error())
				promiseFinish = true
				break
			} else {
				f.SetStatus(FutureStatusFinish)
			}
		}

		if f.isFinish() {
			continue
		}

		break
	}

	if promiseFinish {
		// 检查是否仍然存在于 Manager 中
		if _, exists := p.pm.promises[p.Id]; exists {
			if p.cbFunc != nil {
				p.cbFunc(p.context)
			}
			p.pm.DeletePromise(p.Id)
		}
	}
}

// Update 更新 超时时间，传入秒数
func (pm *Manager) Update(during time.Duration) {
	if during == 0 {
		during = MaxRuntime
	}
	now := time.Now()
	for _, p := range pm.promises {
		d := now.Sub(p.createTime)
		if d >= (time.Second * during) { //(time.Minute * during) {
			p.context.Err = errors.New(fmt.Sprintf("owner[%v] promise[%v][%v] run timeout", pm.owner, p.Id, p.Name))
			if p.cbFunc != nil {
				p.cbFunc(p.context)
			}
			p.pm.DeletePromise(p.Id)
			pm.logger.ErrorF("Manager: Owner[%v] promise[%v][%v] run timeout, delete", pm.owner, p.Id, p.Name)
		}
	}

	if time.Now().Unix()-pm.lastPrintPromiseStat.Unix() >= 5 {
		pm.printStat()
		pm.lastPrintPromiseStat = time.Now()
	}
}

func (pm *Manager) getStat() (time.Duration, time.Duration) {
	statCount := len(pm.statResults)
	if statCount > 0 && pm.maxResult != nil {
		var sumDuration time.Duration
		for _, r := range pm.statResults {
			sumDuration += r.Duration
		}
		avgDuration := time.Duration(sumDuration.Nanoseconds() / int64(statCount))

		return avgDuration, pm.maxResult.Duration
	}

	return time.Duration(0), time.Duration(0)
}

func getDurationDesc(duration time.Duration) string {
	if int(duration.Seconds()) > 0 {
		return fmt.Sprintf("%v second ", duration.Seconds())
	}

	milliseconds := duration.Nanoseconds() / int64(1000*1000)
	if milliseconds > 0 {
		return fmt.Sprintf("%v millisecond ", milliseconds)
	}

	return duration.String()
}

func (pm *Manager) printStat() {
	avgDuration, maxDuration := pm.getStat()
	count := len(pm.statResults)

	if len(pm.statResults) > 0 {
		pm.logger.InfoF("Manager: Owner[%v] PromiseStatus [PromiseCount(%v), AvgPromiseDuration(%v), MaxPromiseDuration(%v) PromiseName(%v)",
			pm.owner, count, getDurationDesc(avgDuration), getDurationDesc(maxDuration), pm.maxResult.Name)
		if pm.warnThresholdFunc != nil {
			thresholdValue := pm.warnThresholdFunc()
			if thresholdValue != 0 && maxDuration >= time.Duration(thresholdValue)*time.Millisecond {
				pm.logger.WarnF("Manager: Owner[%v] PromiseName %v timeout excute %v exceed limit %v",
					pm.owner, pm.maxResult.Name, getDurationDesc(pm.maxResult.Duration), thresholdValue)
			}
		}
	}

	pm.maxResult = nil
	pm.statResults = []*StatResult{}
}

func GetPromiseFutureId(pfId uint64) (promiseId, futureId uint32) {
	promiseId = uint32(pfId >> 32)
	futureId = uint32(pfId)
	return
}

func GetPfId(promiseId, futureId uint32) uint64 {
	pfId := uint64(promiseId) << 32
	pfId += uint64(futureId)

	return pfId
}

func GetStack() string {
	buf := make([]byte, 1<<20)
	l := runtime.Stack(buf, false)
	return string(buf[:l])
}

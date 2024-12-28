package timermgr

import (
	"fmt"
	"github.com/qiafan666/gotato/commons/gapp2/timer"
	"github.com/qiafan666/gotato/commons/gcommon/sval"
	"github.com/qiafan666/gotato/commons/gtime/logictime"
)

// AccType Timer 加速类型
type AccType int32

// AccType
const (
	AccAbs AccType = iota // 按绝对值加速 必须>0
	AccPct                // 按百分比加速 万分比 [0, 10000]
)

// const
const (
	PctBase = 10000 // AccPct 基于万分比
)

// TimerHandler 定时器回调函数类型
type TimerHandler func(int64, sval.M)

// TimerMeta 回调数据
type TimerMeta struct {
	handler TimerHandler
	needDB  bool
}

type dirtyFlag int

const (
	dirtyNew    = 1
	dirtyUpdate = 2
	dirtyDel    = 4
)

// TimerMgr 管理所有的Timer
type TimerMgr struct {
	collName       string
	dbName         string
	Timers         map[int64]*Timer // 需要存盘的Timers，为nil则不能注册需要落地的Timer，TimerMgr本身也就不会落地
	timersUnNeedDB map[int64]*Timer
	dirtyTimers    map[int64]dirtyFlag  // 记录变更过等待同步到DB的Timer(不包含timerUnNeedDB)
	handlers       map[int32]*TimerMeta // 定时器回调函数
	dispatcher     timer.ITimerAPI
}

// NewTimerMgr 通常结合 skeleton module使用
func NewTimerMgr(dispatcher timer.ITimerAPI, dbname, collName string) *TimerMgr {
	return &TimerMgr{
		collName:       collName,
		dbName:         dbname,
		Timers:         make(map[int64]*Timer),
		timersUnNeedDB: make(map[int64]*Timer),
		dirtyTimers:    make(map[int64]dirtyFlag),
		handlers:       make(map[int32]*TimerMeta),
		dispatcher:     dispatcher,
	}
}

// NewTimerMgrNoDB 创建不需要落地的TimerMgr 通常结合 skeleton module使用
func NewTimerMgrNoDB(dispatcher timer.ITimerAPI) *TimerMgr {
	return &TimerMgr{
		Timers:         nil,
		timersUnNeedDB: make(map[int64]*Timer),
		handlers:       make(map[int32]*TimerMeta),
		dispatcher:     dispatcher,
	}
}

// initAfterDB db加载后初始化
func (tm *TimerMgr) initAfterDB() {
	// 恢复timer
	ops := make([]*timer.NewOp, 0, len(tm.Timers))
	for _, t := range tm.Timers {
		ops = append(ops, &timer.NewOp{
			Typ:   t.TimerType,
			ID:    t.TimerID,
			EndTs: t.EndTs,
			Cb:    tm.timerCommonCb,
		})
	}
	tm.dispatcher.BatchNewTimers(ops)
	tm.dispatcher.Logger().DebugF("timermgr initAfterDB: ops %v", len(ops))
}

// 添加一个timer
func (tm *TimerMgr) addTimer(t *Timer, needDB bool) {
	if needDB {
		tm.Timers[t.TimerID] = t
		tm.dirtyTimers[t.TimerID] = dirtyNew
	} else {
		tm.timersUnNeedDB[t.TimerID] = t
	}
}

// 更新一个timer
func (tm *TimerMgr) updateTimer(t *Timer, endTs int64) {
	t.update(endTs)
	if _, ok := tm.Timers[t.TimerID]; ok {
		v := tm.dirtyTimers[t.TimerID]
		tm.dirtyTimers[t.TimerID] = v | dirtyUpdate
	}
}

// 删除一个timer
func (tm *TimerMgr) delTimer(timerID int64) {
	delete(tm.timersUnNeedDB, timerID)
	if _, ok := tm.Timers[timerID]; !ok {
		return
	}
	delete(tm.Timers, timerID)
	if v, ok := tm.dirtyTimers[timerID]; ok {
		if v&dirtyNew != 0 {
			// 如果在保存间隔期间，Timer新建然后删除了，那么Timer本身不需要任何DB操作(相当于没有出现过)
			delete(tm.dirtyTimers, timerID)
		} else {
			tm.dirtyTimers[timerID] = v | dirtyDel
		}
	} else {
		tm.dirtyTimers[timerID] = dirtyDel
	}
}

func (tm *TimerMgr) getTimer(timerID int64) *Timer {
	t := tm.Timers[timerID]
	if t == nil {
		t = tm.timersUnNeedDB[timerID]
	}
	return t
}

func (tm *TimerMgr) timerCommonCb(timerID int64) {
	t := tm.getTimer(timerID)
	if t == nil {
		return
	}
	now := logictime.NowMs()
	if now < t.EndTs {
		tm.dispatcher.Logger().ErrorF("timerCommonCb timer endTs smaller than nowMs")
	}
	timerType := t.TimerType
	f := tm.handlers[timerType].handler
	defer func() {
		if t.IsTicker {
			dura := t.EndTs - t.StartTs
			t.StartTs = now
			t.EndTs = t.StartTs + dura
			tm.dispatcher.NewTimer(timerType, timerID, t.EndTs, tm.timerCommonCb)
		} else {
			tm.delTimer(timerID)
		}
	}()
	f(timerID, t.TimerData)
}

// GetTimer 获取定时器数据
func (tm *TimerMgr) GetTimer(timerID int64) *Timer {
	return tm.getTimer(timerID)
}

// GetTimerByType 根据timer type获取定时器数据，调用这个函数的timer type只有一个timer
func (tm *TimerMgr) GetTimerByType(timerType int32) *Timer {
	// 我假设需要db存储的timer, 一般都有保存timerID
	// 因此这里先在不需要db的timer查找
	for _, t := range tm.timersUnNeedDB {
		if t.TimerType == timerType {
			return t
		}
	}
	for _, t := range tm.Timers {
		if t.TimerType == timerType {
			return t
		}
	}

	return nil
}

func (tm *TimerMgr) newTimer(startTs, endTs int64, timerType int32, timerData sval.M, isTicker bool) int64 {
	h, ok := tm.handlers[timerType]
	if !ok {
		tm.dispatcher.Logger().ErrorF("timer type %v cannot found", timerType)
		return 0
	}
	timerID := tm.dispatcher.NewTimer(timerType, 0, endTs, tm.timerCommonCb)
	t := &Timer{
		TimerID:   timerID,
		TimerType: timerType,
		StartTs:   startTs,
		EndTs:     endTs,
		TimerData: timerData,
		IsTicker:  isTicker,
	}
	tm.addTimer(t, h.needDB)
	return timerID
}

// NewTimerByDura 启动一个Timer 返回TimerID 0为创建失败
func (tm *TimerMgr) NewTimerByDura(duraMs int64, timerType int32, timerData sval.M) int64 {
	now := logictime.NowMs()
	return tm.newTimer(now, now+duraMs, timerType, timerData, false)
}

// NewTimerByTs 返回TimerID 0为创建失败
func (tm *TimerMgr) NewTimerByTs(endTs int64, timerType int32, timerData sval.M) int64 {
	now := logictime.NowMs()
	return tm.newTimer(now, endTs, timerType, timerData, false)
}

// NewTicker 创建一个Ticker，返回TimerID 0为创建失败
// 1. Ticker相当于N次Timer，期间TimerID不会变，使用方可安全保存用于后续随时取消
// 2. Ticker会等本次回调函数执行结束后，再发起下一次Timer，避免Timer堆积
// 3. 回调函数Panic会被defer并且正常发起下一次Timer
func (tm *TimerMgr) NewTicker(duraMs int64, timerType int32, timerData sval.M) int64 {
	now := logictime.NowMs()
	return tm.newTimer(now, now+duraMs, timerType, timerData, true)
}

// AccTimer 加速Timer; micAcc 最少加速时间，AccPct 时生效
func (tm *TimerMgr) AccTimer(timerID int64, accType AccType, value, minAcc int64) (int64, error) {
	nowTs := logictime.NowMs()
	t := tm.getTimer(timerID)
	if t == nil {
		return 0, fmt.Errorf("acc timer failed, timer %v not found", timerID)
	}
	remain := t.EndTs - nowTs
	var newRemain int64
	switch accType {
	case AccAbs:
		if value <= 0 {
			return 0, fmt.Errorf("acc timer failed, invalid args: %d %d %d", timerID, accType, value)
		}
		newRemain = max(0, remain-value)
	case AccPct:
		if value <= 0 && value > PctBase {
			return 0, fmt.Errorf("acc timer failed, invalid args: %d %d %d", timerID, accType, value)
		}
		accTs := max(remain*value/PctBase, minAcc)
		if accTs <= 0 {
			return 0, nil
		}
		newRemain = max(0, remain-accTs)
	default:
		return 0, fmt.Errorf("acc timer failed, invalid args: %d %d %d", timerID, accType, value)
	}
	newEndTs := nowTs + newRemain
	tm.updateTimer(t, newEndTs)
	tm.dispatcher.UpdateTimer(timerID, newEndTs)

	return remain - newRemain, nil
}

// DelayTimer 延迟Timer
func (tm *TimerMgr) DelayTimer(timerID int64, accType AccType, value int64) error {
	nowTs := logictime.NowMs()
	t := tm.getTimer(timerID)
	if t == nil {
		return fmt.Errorf("delay timer failed, timer %v not found", timerID)
	}
	remain := t.EndTs - nowTs
	var newRemain int64
	switch accType {
	case AccAbs:
		if value <= 0 {
			return fmt.Errorf("delay timer failed, invalid args: %d %d %d", timerID, accType, value)
		}
		newRemain = remain + value
	case AccPct:
		if value <= 0 && value > PctBase {
			return fmt.Errorf("delay timer failed, invalid args: %d %d %d", timerID, accType, value)
		}
		newRemain = remain * (PctBase + value) / PctBase
	default:
		return fmt.Errorf("delay timer failed, invalid args: %d %d %d", timerID, accType, value)
	}
	newEndTs := nowTs + newRemain
	tm.updateTimer(t, newEndTs)
	tm.dispatcher.UpdateTimer(timerID, newEndTs)
	return nil
}

// CancelTimer 取消一个定时器
func (tm *TimerMgr) CancelTimer(timerID int64) {
	if timerID == 0 {
		tm.dispatcher.Logger().ErrorF("TimerMgr CancelTimer timerID = 0")
		return
	}
	tm.dispatcher.CancelTimer(timerID)
	tm.delTimer(timerID)
}

// RegisterTimer 注册指定类型timer处理函数
func (tm *TimerMgr) RegisterTimer(timerType int32, handler TimerHandler, needDB bool) {
	if _, ok := tm.handlers[timerType]; ok {
		tm.dispatcher.Logger().ErrorF("TimerMgr RegisterTimer repeat register; timerType:%d", timerType)
	}
	// 如果TimerMgr本身以NoDB模式启动，则不能注册needDB为true的Timer
	if tm.Timers == nil && needDB {
		tm.dispatcher.Logger().ErrorF("TimerMgr no db RegisterTimer need db, timerType:%d", timerType)
	}
	tti := &TimerMeta{
		handler: handler,
		needDB:  needDB,
	}
	tm.handlers[timerType] = tti
}

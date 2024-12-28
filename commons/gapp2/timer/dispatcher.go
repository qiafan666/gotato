package timer

/* Timer 将Timer跑在单独的goroutine中
 * 以精度换取效率
 * 此模块结合Skeleton使用
 */

import (
	"github.com/qiafan666/gotato/commons/gface"
	"github.com/qiafan666/gotato/commons/gid"
	"sort"
	"time"
)

// Dispatcher 定时器分发
// 使用时间轮算法
// 所有的对外API均goroutine safe
type Dispatcher struct {
	timerSlots [timerLevel]map[int64]*Timer // 时间轮 每个level对应的slots中的timer剩余到期时间大于等于 timerTick<<level
	chanOp     chan any                     // 用于向Dispather发送Timer相关操作命令
	nowMs      func() int64
	logger     gface.Logger
}

// newDispatcher nowMs: 外部注入的时间获取接口
func newDispatcher(nowMs func() int64, logger gface.Logger) *Dispatcher {
	disp := new(Dispatcher)
	for k := range disp.timerSlots {
		disp.timerSlots[k] = make(map[int64]*Timer)
	}

	disp.chanOp = make(chan any, _timerOpChanSize)
	disp.nowMs = nowMs
	disp.logger = logger

	return disp
}

// Run 运行分发器
func (disp *Dispatcher) Run() {
	go disp.run()
}

func (disp *Dispatcher) run() {
	defer func() {
		if x := recover(); x != nil {
			disp.logger.ErrorF("TIMER CRASHED %v", x)
		}
	}()

	lastTick := disp.nowMs() / timerTick
	tickTimer := time.NewTimer(timerTick * time.Millisecond)
	for {
		select {
		case t := <-disp.chanOp:
			if t == nil {
				return
			}
			disp.doOp(t)
		case <-tickTimer.C:
			tickTimer.Reset(timerTick * time.Millisecond)
			lastTick = disp.doTick(disp.nowMs(), lastTick)
		}
	}
}

// 删除并返回Timer
func (disp *Dispatcher) delete(timerID int64) *Timer {
	for i := timerLevel - 1; i >= 0; i-- {
		slotMap := disp.timerSlots[i]
		if v, ok := slotMap[timerID]; ok {
			delete(slotMap, timerID)
			return v
		}
	}
	return nil
}

// 将Timer放到合适的时间轮中
func (disp *Dispatcher) place(t *Timer) {
	diff := t.endTs - disp.nowMs()
	if diff < timerTick {
		diff = timerTick
	}
	for i := timerLevel - 1; i >= 0; i-- {
		if diff >= (timerTick << uint(i)) {
			disp.timerSlots[i][t.id] = t
			break
		}
	}
}

func (disp *Dispatcher) doTick(nowMs int64, lastTick0 int64) int64 {
	lastTick := lastTick0
	// 防止服务器时间手动调整前移后 Timer重复触发
	nowTick := nowMs / timerTick
	deltaTick := nowTick - lastTick
	if deltaTick < 1 {
		return nowTick
	}

	maxLevelNeedTrigger := 0 // 如果 level i 被触发，则i-1必然被触发，因此只需记录最大的可触发level
	if deltaTick <= 128 {    // 如果 deltaTick <= 128(大概1秒)，认为是容忍误差，逐步向前tick，但保证每个level至多触发一次
		for ; lastTick < nowTick; lastTick++ {
			for i := timerLevel - 1; i >= 0; i-- {
				mask := (1 << uint(i)) - 1
				if lastTick&int64(mask) == 0 {
					if i > maxLevelNeedTrigger {
						maxLevelNeedTrigger = i
					}
					break
				}
			}
		}
	} else { // 如果 deltaTick 相差太大，即时间被手动调整后移，则直接驱动所有level
		maxLevelNeedTrigger = timerLevel - 1
	}
	// 服务器时间手动调整后移后 Timer向前触发
	for i := maxLevelNeedTrigger; i >= 0; i-- {
		disp.trigger(nowMs, i)
	}
	return nowTick
}

// 单级触发
func (disp *Dispatcher) trigger(nowMs int64, level int) {
	slotMap := disp.timerSlots[level]
	if level != 0 {
		for k, v := range slotMap {
			diff := v.endTs - nowMs
			newLevel := level
			// 直接跳到其对应的槽位中
			for diff < (timerTick<<newLevel) && newLevel > 0 {
				newLevel--
			}
			if newLevel != level {
				// 正常前移到更小的时间轮中
				disp.timerSlots[newLevel][k] = v
				delete(slotMap, k)
			}
		}
		return
	}

	// level == 0 已经是最小的时间轮
	var tmpList timerList
	for _, v := range slotMap {
		if v.endTs < nowMs {
			tmpList = append(tmpList, v)
		}
	}
	if len(tmpList) == 0 {
		return
	}
	// 对待触发的Timer排序
	sort.Sort(tmpList)

	// 开始触发
	for _, t := range tmpList {
		// 本地触发，执行完成即清除Timer
		if t.ownerChan == nil {
			t.Cb()
			delete(slotMap, t.id)
			continue
		}
		// 远程触发，仅当非阻塞写入成功才清除timer
		select { // 发送必须为非阻塞, 传入的chan不能关闭
		case t.ownerChan <- t:
			delete(slotMap, t.id) // 如果发送失败，则尝试下次再次触发
		default:
		}
	}
}

// Stop 停止Dispatcher
func (disp *Dispatcher) Stop() {
	disp.pendOp(nil)
}

// UpdateTimer 加速 Timer
func (disp *Dispatcher) UpdateTimer(timerID, newEndTs int64) {
	if timerID == 0 {
		disp.logger.ErrorF("TimerDispatcher UpdateTimer: timerID == 0")
		return
	}
	disp.pendOp(&UpdateOp{TimerID: timerID, NewEndTs: newEndTs})
}

// NewTimer 创建一个定时器
func (disp *Dispatcher) NewTimer(timerType int32, timerID, timeout int64, cb timerCb, ownerChan chan *Timer) int64 {
	id := timerID
	if id == 0 {
		id = gid.ID()
	}
	newOp := &NewOp{Typ: timerType, ID: id, EndTs: timeout, Cb: cb, OwnerChan: ownerChan}
	disp.pendOp(newOp)
	return id
}

// CancelTimer 取消定时器
func (disp *Dispatcher) CancelTimer(timerID int64) {
	if timerID == 0 {
		disp.logger.ErrorF("TimerDispatcher CancelTimer: timerID == 0")
		return
	}
	disp.pendOp(&CancelOp{TimerID: timerID})
}

func (disp *Dispatcher) BatchNewTimers(ops []*NewOp) []int64 {
	timerIDs := make([]int64, 0, len(ops))
	pOp := &BatchOp{
		Ops: make([]any, 0, len(ops)),
	}

	for _, op := range ops {
		timerID := op.ID
		if timerID == 0 {
			timerID = gid.ID()
		}
		op1 := *op
		op1.ID = timerID
		pOp.Ops = append(pOp.Ops, &op1)
		timerIDs = append(timerIDs, timerID)
	}

	disp.pendOp(pOp)
	return timerIDs
}

func (disp *Dispatcher) pendOp(op any) {
	disp.chanOp <- op
}

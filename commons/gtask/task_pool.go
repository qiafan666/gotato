package gtask

import (
	"errors"
	"github.com/qiafan666/gotato/commons/gface"
	"hash/fnv"
	"runtime"
	"strconv"
	"strings"
)

var defaultPool *Pool

func InitDefaultPool(taskNum, chanNum int, logger gface.Logger) {
	if defaultPool == nil {
		defaultPool = NewTaskPool(taskNum, chanNum, logger)
	}
}

func AddTask(f func(), cb func(), poolDecide string) error {
	if defaultPool == nil {
		return errors.New("task pool is nil")
	}

	defaultPool.AddTask(f, cb, poolDecide)
	return nil
}

func StopDefaultPool() {
	if defaultPool == nil {
		return
	}
	defaultPool.Stop()
}

type UpdateTask struct {
	t chan *taskFuncPair
}

type Pool struct {
	Tasks    []*UpdateTask
	curIndex uint32
	chanNum  int
	stopped  bool
	logger   gface.Logger
}

func NewTaskPool(taskNum, chanNum int, logger gface.Logger) *Pool {
	if logger == nil {
		panic("logger cannot be nil")
	}

	pool := &Pool{
		Tasks:  []*UpdateTask{},
		logger: logger,
	}
	// 有配置初始化，检查下无效的值
	if taskNum <= 0 {
		taskNum = 100
	}
	if chanNum <= 0 {
		chanNum = 10000
	}
	pool.chanNum = chanNum
	for i := 0; i < taskNum; i++ {
		task := &UpdateTask{t: make(chan *taskFuncPair, chanNum)}
		pool.Tasks = append(pool.Tasks, task)
		go ProcessTask(task, logger)
	}

	return pool
}

func (p *Pool) Stop() {
	if p.stopped {
		return
	}

	p.stopped = true

	for _, task := range p.Tasks {
		close(task.t)
	}
}

// AddTask poolDecide 决定固定到指定的 pool 上
func (p *Pool) AddTask(f func(), cb func(), poolDecide string) {
	if len(p.Tasks) == 0 {
		p.logger.ErrorF("pool task len 0")
		return
	}
	var index uint32
	chanAllFull := false

	if poolDecide == "" {
		index = p.curIndex

		// 从当前序号开始找一个未满的 task
		for {
			if len(p.Tasks[int(index)].t) < p.chanNum {
				break
			}
			index = (index + 1) % uint32(len(p.Tasks))

			// 当轮询所有 task 都已经满了后返回
			if index == p.curIndex {
				chanAllFull = true
				break
			}
		}

		// 指向下一个 task 序号
		p.curIndex = (index + 1) % uint32(len(p.Tasks))
	} else {
		// 玩家 ID 固定到对应的 task 上，保证先后
		index = HashString(poolDecide) % uint32(len(p.Tasks))
	}

	t := p.Tasks[int(index)].t
	if len(t) >= p.chanNum {
		_, file, line, _ := runtime.Caller(1)
		ss := strings.Split(file, "/")
		fileName := ss[len(ss)-1]

		id := "[" + fileName + ":" + strconv.Itoa(line) + "] "

		if chanAllFull {
			p.logger.WarnF("add task[%v]. all task is full", id)
		} else {
			p.logger.WarnF("add task[%v]. taskPool index:%d is full", id, index)
		}
	}

	select {
	case t <- &taskFuncPair{
		f:  f,
		cb: cb,
	}:
	default:
		p.logger.ErrorF("task is full")
	}
}

func (t *UpdateTask) executeFun(pair *taskFuncPair, logger gface.Logger) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			l := runtime.Stack(buf, false)
			logger.ErrorF("%v: %s", r, buf[:l])
		}
	}()

	if pair.f != nil {
		pair.f()
	}

	if pair.cb != nil {
		pair.cb()
	}
}

func ProcessTask(task *UpdateTask, logger gface.Logger) {
	if task == nil {
		panic("task is nil")
	}

	for {
		pair, ok := <-task.t
		if !ok {
			break
		}

		task.executeFun(pair, logger)
	}
}

func HashString(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

func (p *Pool) Len() int {
	num := 0
	for _, task := range p.Tasks {
		if task == nil {
			continue
		}

		num += len(task.t)
	}

	return num
}

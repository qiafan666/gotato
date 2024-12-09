package gtask

import (
	"runtime"
	"strconv"
	"strings"
)

const WriteChanNum = 100000

type taskFuncPair struct {
	f  func()
	cb func() // 需要保证线程安全
}

type Task struct {
	functions chan *taskFuncPair
	exitCh    chan int
	logger    taskLogger
}

func New() *Task {
	task := &Task{
		functions: make(chan *taskFuncPair, WriteChanNum),
		exitCh:    make(chan int),
	}

	go func() {
		task.run()
	}()

	return task
}

func (t *Task) AddTask(f func(), cb func()) {
	if len(t.functions) >= WriteChanNum {
		_, file, line, _ := runtime.Caller(1)
		ss := strings.Split(file, "/")
		fileName := ss[len(ss)-1]

		id := "[" + fileName + ":" + strconv.Itoa(line) + "] "
		t.logger.TaskWarnF("add task[%v], but full.", id)
	}

	select {
	case t.functions <- &taskFuncPair{
		f:  f,
		cb: cb,
	}:

	default:
		t.logger.TaskErrorF("task is full")
	}
}

func (t *Task) executeFunc(pair *taskFuncPair) {
	if pair == nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			l := runtime.Stack(buf, false)
			t.logger.TaskErrorF("%v: %s", r, buf[:l])
		}
	}()

	if pair.f != nil {
		pair.f()
	}

	if pair.cb != nil {
		pair.cb()
	}
}

func (t *Task) run() {
	for {
		pair, ok := <-t.functions
		if !ok {
			t.logger.TaskWarnF("task.functions closed")
			break
		}
		t.executeFunc(pair)
	}

	t.exitCh <- 1
}

func (t *Task) Stop() {
	close(t.functions)
	<-t.exitCh
}

func (t *Task) Len() int {
	return len(t.functions)
}

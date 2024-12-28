package timer

type NewOp struct {
	Typ   int32   // 做消息统计用
	ID    int64   // ID
	EndTs int64   // 到期时间戳 ms
	Cb    timerCb // timer 回调
	// timer 触发后会写入该channel，由该channel的持有者执行回调
	// 如果OwnerChan为nil，则直接在Dispatcher中执行回调
	OwnerChan chan *Timer
}

type UpdateOp struct {
	TimerID  int64
	NewEndTs int64
}

type CancelOp struct {
	TimerID int64
}

type BatchOp struct {
	Ops []any
}

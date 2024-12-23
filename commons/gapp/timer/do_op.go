package timer

// 执行对应的操作 返回是否停止Dispatcher
func (disp *Dispatcher) doOp(op any) {
	switch o := op.(type) {
	case *NewOp:
		disp.doNewOp(o)
	case *UpdateOp:
		disp.doUpdateOp(o)
	case *CancelOp:
		disp.doCancelOp(o)
	case *BatchOp:
		disp.doBatchOp(o)
	default:
		disp.logger.ErrorF("unknown type of op: %v", op)
	}
}

func (disp *Dispatcher) doNewOp(op *NewOp) {
	t := &Timer{
		typ:       op.Typ,
		id:        op.ID,
		endTs:     op.EndTs,
		cb:        op.Cb,
		ownerChan: op.OwnerChan,
		logger:    disp.logger,
	}
	disp.place(t)
}

func (disp *Dispatcher) doUpdateOp(op *UpdateOp) {
	// 找到并删除Timer
	old := disp.delete(op.TimerID)
	if old == nil {
		disp.logger.ErrorF("update TimerId=%d, not found", op.TimerID)
		return
	}

	// 重新找合适的框
	old.endTs = op.NewEndTs
	disp.place(old)
}

func (disp *Dispatcher) doCancelOp(op *CancelOp) {
	disp.delete(op.TimerID)
}

func (disp *Dispatcher) doBatchOp(op *BatchOp) {
	for _, childOp := range op.Ops {
		disp.doOp(childOp)
	}
}

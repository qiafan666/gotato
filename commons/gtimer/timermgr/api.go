package timermgr

import (
	"github.com/qiafan666/gotato/commons/gcommon/sval"
)

type API interface {
	RegisterTimer(timerType int32, handler TimerHandler, needDB bool)
	NewTimerByTs(endTs int64, timerType int32, timerData sval.M) int64
	NewTicker(duraMs int64, timerType int32, timerData sval.M) int64
	CancelTimer(timerID int64)
	GetTimer(timerID int64) *Timer
	AccTimer(timerID int64, accType AccType, value, minAcc int64) (int64, error)
	DelayTimer(timerID int64, accType AccType, value int64) (err error)
	NewTimerByDura(duraMs int64, timerType int32, timerData sval.M) int64
}

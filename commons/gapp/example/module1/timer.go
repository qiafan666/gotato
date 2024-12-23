package module1

import (
	"github.com/qiafan666/gotato/commons/gapp/timer/timermgr"
	"github.com/qiafan666/gotato/commons/gcommon/sval"
	"github.com/qiafan666/gotato/commons/gtime/logictime"
	"log"
)

func (m *Module1) initTimer() {
	m.timerMgr = timermgr.NewTimerMgrNoDB(m.skeleton.TimerAPI())

	m.timerMgr.RegisterTimer(1, m.timer1Handler, false)
	m.timerMgr.NewTicker(logictime.Second*1, 1, sval.M{})
}

func (m *Module1) timer1Handler(i int64, m2 sval.M) {
	log.Printf("timer1Handler: %d", i)
}

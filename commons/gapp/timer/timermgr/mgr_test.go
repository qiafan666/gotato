package timermgr

import (
	"github.com/qiafan666/gotato/commons/gapp/timer"
	"github.com/qiafan666/gotato/commons/gcommon/sval"
	"github.com/qiafan666/gotato/commons/gface"
	"github.com/qiafan666/gotato/commons/gtime/logictime"
	"log"
	"testing"
	"time"
)

const (
	TimerTypeTest = 1
)

func TestDispatcher(t *testing.T) {

	// 启动全局定时器逻辑
	timer.Run(func() int64 { return logictime.NowMs() }, gface.NewLogger("timer", nil))

	// 初始化逻辑代理，只实例化一次
	logicDelegate := timer.NewLogicDelegate()

	// 启动调度器主循环
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case ct := <-logicDelegate.ChanTimer(): // 使用单一逻辑代理
				logicDelegate.Exec(ct) // 执行定时器任务
			case <-ticker.C:
				// 调试日志：每 100ms 确保调度器还在运行
				//log.Printf("Main scheduler running...")
			}
		}
	}()

	// 初始化定时器管理器
	timerMgr := NewTimerMgrNoDB(logicDelegate)

	// 注册定时器任务
	timerMgr.RegisterTimer(TimerTypeTest, handleTimer, false)

	// 创建一个定时任务
	timerId := timerMgr.NewTicker(logictime.Second*2, TimerTypeTest, nil)
	log.Printf("Timer created: timerId=%d", timerId)

	go func() {
		// 等待 5s 后停止定时器
		time.Sleep(time.Second * 5)
		timerMgr.CancelTimer(timerId)
	}()

	// 等待任务执行
	time.Sleep(time.Second * 11)

	// 停止全局定时器
	timer.Stop()
}

func handleTimer(id int64, m sval.M) {

	log.Printf("handleTimer id:%d,m:%v", id, m)
}

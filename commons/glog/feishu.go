package glog

import (
	"github.com/qiafan666/gotato/commons/gcast"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/gjson"
	"go.uber.org/zap/zapcore"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	FeiShuLogCountPerSec    = 3    // 每秒3条日志, 飞书限制每秒50条，但有多个server
	FeiShuSameCallerTick    = 1800 // 相同caller的log，每半个小时push一次飞书
	FeiShuRegisterEntryFunc func(zapcore.Entry) error
)

// FeiShuHook 封装了飞书消息发送逻辑和状态管理
type FeiShuHook struct {
	msgChan  chan []byte
	groupID  string
	url      string
	accCount int
	accSec   int64
	stats    map[string]int64
	stop     atomic.Bool
	wg       sync.WaitGroup
}

// NewFeiShuHook 初始化 FeiShuHook
func NewFeiShuHook(groupID, url string) *FeiShuHook {
	hook := &FeiShuHook{
		msgChan: make(chan []byte, 1000),
		groupID: groupID,
		url:     url,
		stats:   make(map[string]int64),
	}
	hook.wg.Add(1)
	go hook.worker()
	return hook
}

// worker 负责处理消息队列并发送日志
func (hook *FeiShuHook) worker() {
	defer hook.wg.Done()
	for msg := range hook.msgChan {
		_, _, _ = gcommon.ProxyRequest(http.MethodPost, http.Header{"Content-Type": {"application/json"}}, hook.url, msg)
	}
}

// Close 停止 FeiShuHook 的工作
func (hook *FeiShuHook) Close() {
	close(hook.msgChan)
	hook.stop.Store(true)
	hook.wg.Wait()
}

// DefaultRegisterHook 默认注册日志钩子，发送日志到 FeiShu
func (hook *FeiShuHook) DefaultRegisterHook() {
	FeiShuRegisterEntryFunc = func(entry zapcore.Entry) error {
		if entry.Level < zapcore.ErrorLevel {
			return nil
		}
		if hook.stop.Load() {
			return nil
		}

		// 控制日志发送频率
		if hook.groupID == "" {
			return nil
		}

		nowSec := entry.Time.Unix()
		caller := entry.Caller.String()
		pushSec := hook.stats[caller]
		if pushSec+gcast.ToInt64(FeiShuSameCallerTick) > nowSec {
			return nil
		}

		// 重置计数器
		if hook.accSec != nowSec {
			hook.accSec = nowSec
			hook.accCount = 0
		}
		if hook.accCount >= FeiShuLogCountPerSec {
			return nil
		}

		hook.stats[caller] = nowSec
		hook.accCount++

		ip, _ := gcommon.GetLocalIP()

		// 组装飞书消息
		sb := strings.Builder{}
		sb.WriteString(gcommon.Kv2Str("",
			"time", entry.Time.Format(time.DateTime),
			"level", entry.Level.String(),
			"caller", caller,
			"message", entry.Message,
			"stack", entry.Stack,
			"ip", ip))

		payload := map[string]string{
			"group_id": hook.groupID,
			"words":    sb.String(),
		}
		b, _ := gjson.Marshal(payload)

		select {
		case hook.msgChan <- b:
		default:
		}
		return nil
	}
}

// CustomRegisterHook 自定义注册日志钩子，发送日志到 FeiShu
func (hook *FeiShuHook) CustomRegisterHook(f func(entry zapcore.Entry) error) error {
	FeiShuRegisterEntryFunc = f
	return nil
}

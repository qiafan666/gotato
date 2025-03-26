package module

import (
	"github.com/qiafan666/gotato/commons/gapp/chanrpc"
	g "github.com/qiafan666/gotato/commons/gapp/go"
	"github.com/qiafan666/gotato/commons/gapp/stat"
	"github.com/qiafan666/gotato/commons/gapp/timer"
	"github.com/qiafan666/gotato/commons/gface"
	"time"
)

// ISkeleton 骨架接口
type ISkeleton interface {
	// TimerAPI 定时器
	TimerAPI() timer.ITimerAPI

	// Run 生命周期
	Run(closeSig chan bool)

	// SafeGo 异步执行
	SafeGo(f, cb func())

	// Server 消息交互
	Server() chanrpc.IServer
	Client() chanrpc.IClient

	// MsgStat 消息状态统计
	MsgStat() map[string]string

	// Logger 日志
	Logger() gface.ILogger
}

// skeleton 模块基础框架
type skeleton struct {
	*g.Go

	timerDelegate timer.ITimerDelegate

	chanCli *chanrpc.Client
	chanSrv *chanrpc.Server
	// cb运行时间统计
	stat *stat.MsgStat[string]

	logger gface.ILogger
}

// NewSkeleton .
func NewSkeleton(goLen, chanrpcLen, asyncCallLen int, logger gface.ILogger) ISkeleton {
	if goLen <= 0 || chanrpcLen < 0 || asyncCallLen < 0 || logger == nil {
		panic("invalid skeleton args")
	}

	s := &skeleton{
		timerDelegate: timer.NewLogicDelegate(),
		chanSrv:       chanrpc.NewServer(chanrpcLen),
		chanCli:       chanrpc.NewClient(asyncCallLen),
		Go:            g.New(goLen),
		stat:          stat.NewStat[string](),
		logger:        logger,
	}

	return s
}

// Run 启动初始化
func (s *skeleton) Run(closeSig chan bool) {
	for {
		select {
		case <-closeSig:
			s.close()
			return
		case ackCtx := <-s.chanCli.ChanAck():
			ts1 := time.Now().UnixMicro()
			s.chanCli.Exec(ackCtx)
			if s.stat != nil {
				cost := time.Now().UnixMicro() - ts1
				s.stat.Add(ackCtx.GetStatName(), cost)
			}
		case reqCtx := <-s.chanSrv.ChanReq():
			ts1 := time.Now().UnixMicro()
			s.chanSrv.Exec(reqCtx)
			if s.stat != nil {
				cost := time.Now().UnixMicro() - ts1
				s.stat.Add(reqCtx.GetStatName(), cost)
				if cost > 300000 { // 大于300毫秒的warn log
					s.logger.DebugF(nil, "skeleton exec too long cost:%v stat name:%s len:%v", cost, reqCtx.GetStatName(), s.chanSrv.Len())
				}
			}
		case cb := <-s.Go.ChanCb:
			s.Go.Cb(cb)
		case t := <-s.timerDelegate.ChanTimer():
			ts1 := time.Now().UnixMicro()
			s.timerDelegate.Exec(t)
			if s.stat != nil {
				cost := time.Now().UnixMicro() - ts1
				s.stat.Add(t.GetStatName(), cost)
			}
		}
	}
}

func (s *skeleton) close() {
	s.chanSrv.Close()
	s.Go.Close()
	s.chanCli.Close()
	if s.stat != nil {
		s.logger.InfoF(nil, "skeleton stat: %v", s.stat.Statistic())
	}
}

// MsgStat 获取前n个处理耗时最长的消息
func (s *skeleton) MsgStat() map[string]string {
	if s.stat == nil {
		return nil
	}
	return s.stat.Statistic()
}

// Server 返回chanrpc Server
func (s *skeleton) Server() chanrpc.IServer {
	return s.chanSrv
}

// Client 返回chanrpc client
func (s *skeleton) Client() chanrpc.IClient {
	return s.chanCli
}

// TimerAPI .
func (s *skeleton) TimerAPI() timer.ITimerAPI {
	return s.timerDelegate
}

// Logger .
func (s *skeleton) Logger() gface.ILogger {
	return s.logger
}

package main

import (
	"context"
	"fmt"
	"github.com/qiafan666/gotato/commons/gapp"
	"github.com/qiafan666/gotato/commons/gapp/chanrpc"
	"github.com/qiafan666/gotato/commons/gapp/example/def"
	"github.com/qiafan666/gotato/commons/gapp/example/module1"
	"github.com/qiafan666/gotato/commons/gapp/example/module2"
	"github.com/qiafan666/gotato/commons/gapp/example/module3"
	"github.com/qiafan666/gotato/commons/gapp/timer"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/gcommon/sval"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

type logger struct{}

func (l *logger) ErrorF(ctx context.Context, format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[ERROR] [%s] ", l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	} else {
		l.Logger().Errorf(fmt.Sprintf(l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	}
}
func (l *logger) WarnF(ctx context.Context, format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[WARN] [%s] ", l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	} else {
		l.Logger().Warnf(fmt.Sprintf(l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	}
}
func (l *logger) InfoF(ctx context.Context, format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[INFO] [%s] ", l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	} else {
		l.Logger().Infof(fmt.Sprintf(l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	}
}
func (l *logger) DebugF(ctx context.Context, format string, args ...interface{}) {
	if l.Logger() == nil {
		log.Printf(fmt.Sprintf("[DEBUG] [%s] ", l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	} else {
		l.Logger().Debugf(fmt.Sprintf(l.Prefix())+gcommon.GetTraceId(ctx)+format, args...)
	}
}
func (l *logger) Logger() *zap.SugaredLogger {
	return def.ZapLog
}
func (l *logger) Prefix() string {
	return "gapp"
}

func main() {
	fmt.Println("test start")
	zapLog()
	timer.Run(nil, &logger{})
	m1 := module1.NewModule()
	m2 := module2.NewModule()
	m3 := module3.NewModule()
	gapp.DefaultApp().Start(&logger{}, m1, m2, m3)

	// m1.ChanSrv().Cast(&iproto.Test1Ntf{PlayerID: 111, Name: "ning1", T1: []int64{1, 2, 3}})

	// 异步消息
	m2.Cast(gcommon.SetTraceId("m2 cast ctx"), def.TEST1, &def.Test1Ntf{PlayerID: 111, Name: "ning1", T1: []int64{1, 2, 3}})

	ackCtx := m2.Call(gcommon.SetTraceId("m2 call ctx"), def.TEST1, &def.Test1Req{PlayerID: 222, Name: "ning2", T1: []int64{2, 3, 4}})
	if ackCtx.Err != nil {
		m2.Logger().ErrorF(gcommon.SetTraceId("m2 call ctx"), "call err:%v", ackCtx.Err)
	} else {
		ack := ackCtx.Ack.(*def.Test1Ack)
		m2.Logger().InfoF(gcommon.SetTraceId("m2 call result"), "call ret:%v", ack)
	}
	// 异步回调
	m2.AsyncCall(gcommon.SetTraceId("m2 AsyncCall ctx"), def.TEST1, &def.Test1Req{
		PlayerID: 222,
		Name:     "ning2",
		T1:       []int64{2, 3, 4},
	}, func(ackCtx *chanrpc.AckCtx) {
		if ackCtx.Err != nil {
			return
		}
		ack := ackCtx.Ack.(*def.Test1Ack)
		m2.Logger().InfoF(gcommon.SetTraceId("m2 AsyncCall ctx"), "async call:%+v", ack)
	}, nil)

	// 异步回调带上下文
	m2.AsyncCall(gcommon.SetTraceId("m2 AsyncCall ctx"), def.TEST1, &def.Test1Req{
		PlayerID: 222,
		Name:     "ning2",
		T1:       []int64{3, 4, 5},
	}, func(ackCtx *chanrpc.AckCtx) {
		if ackCtx.Err != nil {
			return
		}
		ack := ackCtx.Ack.(*def.Test1Ack)
		m2.Logger().WarnF(gcommon.SetTraceId("m2 AsyncCall ctx"), "async call with ctx:%+v %+v", ack, ackCtx.M)
	}, sval.M{"111": sval.Int64(4444)})

	// 同步调用
	ret := m2.Call(gcommon.SetTraceId("m2 call ctx"), def.TEST1, &def.Test1CallReq{PlayerID: 333, Name: "ning3", T1: []int64{3, 4, 5}})
	if ret.Err != nil {
		log.Printf("call err:%v", ret.Err)
	} else {
		ack := ret.Ack.(*def.Test1CallAck)
		log.Printf("call ret:%+v", ack)
	}

	// 同步调用actor
	actorRet := m2.CallActor(gcommon.SetTraceId("m2 callActor"), def.TEST3, 111, &def.Test1ActorReq{PlayerID: 444, Name: "ning4", T1: []int64{4, 5, 6}})
	if actorRet.Err != nil {
		log.Printf("call actor err:%v", actorRet.Err)
	} else {
		ack := actorRet.Ack.(*def.Test1ActorAck)

		log.Printf("call actor ret:%+v", ack)
	}
	time.Sleep(3 * time.Second)
	timer.Stop()
	fmt.Println("test end")
}

func zapLog() {

	writeSyncer := getLogWriter(fmt.Sprintf("%s/%s.log", "./log", "test"))

	encoder := DevEncoder()

	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)

	// zap.AddCaller()  添加将调用函数信息记录到日志中的功能。
	def.ZapLog = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
}

func getLogWriter(logPath string) zapcore.WriteSyncer {
	dir := path.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create log directory: %s", err))
	}
	return zapcore.AddSync(io.MultiWriter(&lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     1, //days
		LocalTime:  true,
		Compress:   true, // disabled by default
	}, os.Stdout))
}

// SimpleEncoder 自定义日志格式 生产环境用这个
func SimpleEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.LineEnding = zapcore.DefaultLineEnding
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// DevEncoder 自定义日志格式 开发测试环境用这个
func DevEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.LineEnding = zapcore.DefaultLineEnding

	// 自定义 EncodeCaller 方法，提取方法名
	encoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		if caller.Defined {
			// 提取文件路径和行号
			fileWithLine := fmt.Sprintf("%s:%d", caller.File, caller.Line)

			// 提取最后两级路径部分
			pathParts := strings.Split(fileWithLine, "/")
			if len(pathParts) > 2 {
				fileWithLine = strings.Join(pathParts[len(pathParts)-2:], "/")
			}

			// 提取方法名
			funcName := runtime.FuncForPC(caller.PC).Name()
			if funcName != "" {
				// 分割并提取最后一级方法名
				split := strings.Split(funcName, ".")
				if len(split) > 2 {
					if strings.HasPrefix(split[len(split)-1], "func") {
						funcName = split[len(split)-2]
					} else {
						funcName = split[len(split)-1] // 只有方法名
					}
				} else if len(split) == 2 {
					funcName = split[1] // 只有方法名
				} else if len(split) == 1 {
					funcName = split[0] // 只有方法名
				} else {
					funcName = "unknown"
				}
			} else {
				funcName = "unknown"
			}

			// 输出格式为 文件:行号 [方法名]，避免重复
			enc.AppendString(fmt.Sprintf("%s   [function_name:%s]", fileWithLine, funcName))
		} else {
			enc.AppendString("unknown")
		}
	}

	return zapcore.NewConsoleEncoder(encoderConfig)
}

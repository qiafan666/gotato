package glog

import (
	"context"
	"fmt"
	"github.com/qiafan666/gotato/commons"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/gconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

var Slog Logger
var Gorm GormLogger
var ZapLog *zap.SugaredLogger
var GormLog *zap.SugaredLogger

// GormSkip gorm日志调用函数栈的层数, 默认为5
var GormSkip = 5

// GormSlowSqlDuration 慢sql日志打印的阈值
var GormSlowSqlDuration = time.Second * 3

type Logger struct {
}

func init() {
	gormEncoder := getGormEncoder()
	logEncoder := getLogEncoder()

	Slog = Logger{}
	Gorm = GormLogger{
		LogLevel:                  commons.LogLevel[gconfig.SC.SConfigure.GormLogLevel],
		IgnoreRecordNotFoundError: true,
		SlowSqlTime:               GormSlowSqlDuration,
	}

	writeSyncer := getLogWriter(fmt.Sprintf("%s/%s.log", gconfig.SC.SConfigure.LogPath, gconfig.SC.SConfigure.LogName))
	gormCore := zapcore.NewCore(gormEncoder, writeSyncer, commons.ZapLogLevel[gconfig.SC.SConfigure.ZapLogLevel])
	logCore := zapcore.NewCore(logEncoder, writeSyncer, commons.ZapLogLevel[gconfig.SC.SConfigure.ZapLogLevel])

	if GormLog != nil {
		_ = GormLog.Sync()
	}
	if ZapLog != nil {
		_ = ZapLog.Sync()
	}
	// zap.AddCaller()  添加将调用函数信息记录到日志中的功能。
	GormLog = zap.New(gormCore, zap.AddCaller(), zap.AddCallerSkip(5)).Sugar()
	ZapLog = zap.New(logCore, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()

}

func ReInit() {
	gormEncoder := getGormEncoder()
	logEncoder := getLogEncoder()

	Slog = Logger{}
	Gorm = GormLogger{
		LogLevel:                  commons.LogLevel[gconfig.SC.SConfigure.GormLogLevel],
		IgnoreRecordNotFoundError: true,
		SlowSqlTime:               GormSlowSqlDuration,
	}
	writeSyncer := getLogWriter(fmt.Sprintf("%s/%s.log", gconfig.SC.SConfigure.LogPath, gconfig.SC.SConfigure.LogName))
	gormCore := zapcore.NewCore(gormEncoder, writeSyncer, commons.ZapLogLevel[gconfig.SC.SConfigure.ZapLogLevel])
	logCore := zapcore.NewCore(logEncoder, writeSyncer, commons.ZapLogLevel[gconfig.SC.SConfigure.ZapLogLevel])

	if GormLog != nil {
		_ = GormLog.Sync()
	}
	if ZapLog != nil {
		_ = ZapLog.Sync()
	}
	// zap.AddCaller()  添加将调用函数信息记录到日志中的功能。
	GormLog = zap.New(gormCore, zap.AddCaller(), zap.AddCallerSkip(GormSkip)).Sugar()
	ZapLog = zap.New(logCore, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()

}

func getGormEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.LineEnding = zapcore.DefaultLineEnding
	return zapcore.NewConsoleEncoder(encoderConfig)
}
func getLogEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.LineEnding = zapcore.DefaultLineEnding
	encoderConfig.FunctionKey = "func"

	// 自定义 EncodeCaller 方法，提取方法名
	encoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		if caller.Defined {
			// 提取文件路径和行号
			fileWithLine := fmt.Sprintf("%s:%d", caller.File, caller.Line)

			// 提取方法名
			funcName := runtime.FuncForPC(caller.PC).Name()
			if funcName != "" {
				// 去掉包路径，仅保留方法名
				lastSlash := strings.LastIndex(funcName, "/")
				if lastSlash != -1 {
					funcName = funcName[lastSlash+1:] // 去掉路径部分
				}
				lastDot := strings.LastIndex(funcName, ".")
				if lastDot != -1 {
					funcName = funcName[lastDot+1:] // 去掉包名部分
				}
			} else {
				funcName = ""
			}

			// 组合路径和方法名
			enc.AppendString(fmt.Sprintf("%s [%s]", fileWithLine, funcName))
		} else {
			enc.AppendString("")
		}
	}

	return zapcore.NewConsoleEncoder(encoderConfig)
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

func GetTraceId(ctx context.Context) string {
	if traceId, ok := ctx.Value("trace_id").(string); ok {
		return fmt.Sprintf("【trace_id:%s】", traceId)
	} else {
		return ""
	}
}

func SetTraceId(traceId string) context.Context {
	return context.WithValue(context.Background(), "trace_id", traceId)
}

func SetTraceIdWithCtx(ctx context.Context, traceId string) context.Context {
	return context.WithValue(ctx, "trace_id", traceId)
}

func (l *Logger) InfoF(ctx context.Context, template string, args ...interface{}) {
	if ctx != nil {
		ZapLog.Infof(GetTraceId(ctx)+template, args...)
	} else {
		ZapLog.Infof(template, args...)
	}
}

func (l *Logger) InfoKVs(ctx context.Context, msg string, kv ...any) {

	if ctx != nil {
		ZapLog.Infof(GetTraceId(ctx) + gcommon.Kv2String(msg, kv...))
	} else {
		ZapLog.Infof(gcommon.Kv2String(msg, kv...))
	}
}

func (l *Logger) DebugF(ctx context.Context, template string, args ...interface{}) {
	if ctx != nil {
		ZapLog.Debugf(GetTraceId(ctx)+template, args...)
	} else {
		ZapLog.Debugf(template, args...)
	}
}

func (l *Logger) DebugKVs(ctx context.Context, msg string, kv ...any) {
	if ctx != nil {
		ZapLog.Debugf(GetTraceId(ctx) + gcommon.Kv2String(msg, kv...))
	} else {
		ZapLog.Debugf(gcommon.Kv2String(msg, kv...))
	}
}

func (l *Logger) WarnF(ctx context.Context, template string, args ...interface{}) {
	if ctx != nil {
		ZapLog.Warnf(GetTraceId(ctx)+template, args...)
	} else {
		ZapLog.Warnf(template, args...)
	}
}

func (l *Logger) WarnKVs(ctx context.Context, msg string, kv ...any) {
	if ctx != nil {
		ZapLog.Warnf(GetTraceId(ctx) + gcommon.Kv2String(msg, kv...))
	} else {
		ZapLog.Warnf(gcommon.Kv2String(msg, kv...))
	}
}

func (l *Logger) ErrorF(ctx context.Context, template string, args ...interface{}) {
	if ctx != nil {
		ZapLog.Errorf(GetTraceId(ctx)+template, args...)
	} else {
		ZapLog.Errorf(template, args...)
	}
}

func (l *Logger) ErrorKVs(ctx context.Context, msg string, kv ...any) {
	if ctx != nil {
		ZapLog.Errorf(GetTraceId(ctx) + gcommon.Kv2String(msg, kv...))
	} else {
		ZapLog.Errorf(gcommon.Kv2String(msg, kv...))
	}
}

func (l *Logger) PanicF(ctx context.Context, template string, args ...interface{}) {
	if ctx != nil {
		ZapLog.Panicf(GetTraceId(ctx)+template, args...)
	} else {
		ZapLog.Panicf(template, args...)
	}
}

func (l *Logger) PanicKVs(ctx context.Context, msg string, kv ...any) {
	if ctx != nil {
		ZapLog.Panicf(GetTraceId(ctx) + gcommon.Kv2String(msg, kv...))
	} else {
		ZapLog.Panicf(gcommon.Kv2String(msg, kv...))
	}
}

func (l *Logger) Printf(format string, v ...interface{}) {
	ZapLog.Infof(format, v...)
}

func (l *Logger) PrintKvs(msg string, kv ...any) {
	ZapLog.Info(gcommon.Kv2String(msg, kv...))
}

func (l *Logger) Print(v ...interface{}) {
	ZapLog.Info(v...)
}

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
	encoder := getEncoder()
	Slog = Logger{}
	Gorm = GormLogger{
		LogLevel:                  commons.LogLevel[gconfig.SC.SConfigure.GormLogLevel],
		IgnoreRecordNotFoundError: true,
		SlowSqlTime:               GormSlowSqlDuration,
	}
	writeSyncer := getLogWriter(fmt.Sprintf("%s/%s.log", gconfig.SC.SConfigure.LogPath, gconfig.SC.SConfigure.LogName))
	core := zapcore.NewCore(encoder, writeSyncer, commons.ZapLogLevel[gconfig.SC.SConfigure.ZapLogLevel])
	// zap.AddCaller()  添加将调用函数信息记录到日志中的功能。
	ZapLog = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
	GormLog = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(5)).Sugar()
}

func ReInit() {
	encoder := getEncoder()
	Slog = Logger{}
	Gorm = GormLogger{
		LogLevel:                  commons.LogLevel[gconfig.SC.SConfigure.GormLogLevel],
		IgnoreRecordNotFoundError: true,
		SlowSqlTime:               GormSlowSqlDuration,
	}
	writeSyncer := getLogWriter(fmt.Sprintf("%s/%s.log", gconfig.SC.SConfigure.LogPath, gconfig.SC.SConfigure.LogName))
	core := zapcore.NewCore(encoder, writeSyncer, commons.ZapLogLevel[gconfig.SC.SConfigure.ZapLogLevel])
	// zap.AddCaller()  添加将调用函数信息记录到日志中的功能。
	ZapLog = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
	GormLog = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(GormSkip)).Sugar()
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.LineEnding = zapcore.DefaultLineEnding
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter(logPath string) zapcore.WriteSyncer {
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

func SetTraceIdWithCtx(traceId string, ctx context.Context) context.Context {
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

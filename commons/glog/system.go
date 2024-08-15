package glog

import (
	"context"
	"fmt"
	"github.com/qiafan666/gotato/commons"
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
func getTraceId(ctx context.Context) string {
	if traceId, ok := ctx.Value("trace_id").(string); ok {
		return fmt.Sprintf("【trace_id:%s】", traceId)
	} else {
		return ""
	}
}

func SetTraceId(traceId string) context.Context {
	return context.WithValue(context.Background(), "trace_id", traceId)
}

func (l *Logger) InfoF(ctx context.Context, template string, args ...interface{}) {
	if ctx != nil {
		ZapLog.Infof(getTraceId(ctx)+template, args...)
	} else {
		ZapLog.Infof(template, args...)
	}
}

func (l *Logger) DebugF(ctx context.Context, template string, args ...interface{}) {
	if ctx != nil {
		ZapLog.Debugf(getTraceId(ctx)+template, args...)
	} else {
		ZapLog.Debugf(template, args...)
	}
}

func (l *Logger) WarnF(ctx context.Context, template string, args ...interface{}) {
	if ctx != nil {
		ZapLog.Warnf(getTraceId(ctx)+template, args...)
	} else {
		ZapLog.Warnf(template, args...)
	}
}

func (l *Logger) ErrorF(ctx context.Context, template string, args ...interface{}) {
	if ctx != nil {
		ZapLog.Errorf(getTraceId(ctx)+template, args...)
	} else {
		ZapLog.Errorf(template, args...)
	}
}

func (l *Logger) PanicF(ctx context.Context, template string, args ...interface{}) {
	if ctx != nil {
		ZapLog.Panicf(getTraceId(ctx)+template, args...)
	} else {
		ZapLog.Panicf(template, args...)
	}
}

func (l *Logger) Printf(format string, v ...interface{}) {
	ZapLog.Infof(format, v...)
}
func (l *Logger) Print(v ...interface{}) {
	ZapLog.Info(v...)
}

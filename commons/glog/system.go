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
var FeiShu *FeiShuHook

// GormSkip gorm日志调用函数栈的层数, 默认为5
var GormSkip = 5

// GormSlowSqlDuration 慢sql日志打印的阈值
var GormSlowSqlDuration = time.Second * 3

type Logger struct {
}

func init() {
	Slog = Logger{}
	Gorm = GormLogger{
		LogLevel:                  commons.LogLevel[gconfig.SC.SConfigure.GormLogLevel],
		IgnoreRecordNotFoundError: true,
		SlowSqlTime:               GormSlowSqlDuration,
	}

	writeSyncer := getLogWriter(fmt.Sprintf("%s/%s.log", gconfig.SC.SConfigure.LogPath, gconfig.SC.SConfigure.LogName))

	var encoder zapcore.Encoder
	if gconfig.SC.SConfigure.FunctionName {
		encoder = DevEncoder()
	} else {
		encoder = SimpleEncoder()
	}
	core := zapcore.NewCore(encoder, writeSyncer, commons.ZapLogLevel[gconfig.SC.SConfigure.ZapLogLevel])

	if gconfig.SC.FeiShuConfig.Enable {
		FeiShu = NewFeiShuHook(gconfig.SC.FeiShuConfig.Url, gconfig.SC.FeiShuConfig.GroupId)
		zapcore.RegisterHooks(core, FeiShuRegisterEntryFunc)
	}

	// zap.AddCaller()  添加将调用函数信息记录到日志中的功能。
	GormLog = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(5)).Sugar()
	ZapLog = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
}

func ReInit() {
	Slog = Logger{}
	Gorm = GormLogger{
		LogLevel:                  commons.LogLevel[gconfig.SC.SConfigure.GormLogLevel],
		IgnoreRecordNotFoundError: true,
		SlowSqlTime:               GormSlowSqlDuration,
	}
	writeSyncer := getLogWriter(fmt.Sprintf("%s/%s.log", gconfig.SC.SConfigure.LogPath, gconfig.SC.SConfigure.LogName))
	var encoder zapcore.Encoder
	if gconfig.SC.SConfigure.FunctionName {
		encoder = DevEncoder()
	} else {
		encoder = SimpleEncoder()
	}
	core := zapcore.NewCore(encoder, writeSyncer, commons.ZapLogLevel[gconfig.SC.SConfigure.ZapLogLevel])

	if gconfig.SC.FeiShuConfig.Enable {
		FeiShu = NewFeiShuHook(gconfig.SC.FeiShuConfig.Url, gconfig.SC.FeiShuConfig.GroupId)
		zapcore.RegisterHooks(core, FeiShuRegisterEntryFunc)
	}

	// zap.AddCaller()  添加将调用函数信息记录到日志中的功能。
	GormLog = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(GormSkip)).Sugar()
	ZapLog = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
}

// SimpleEncoder 自定义日志格式 生产环境用这个
func SimpleEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// DevEncoder 自定义日志格式 开发测试环境用这个
func DevEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
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

func (l *Logger) DebugF(ctx context.Context, template string, args ...interface{}) {
	if ctx != nil {
		ZapLog.Debugf(gcommon.GetTraceId(ctx)+template, args...)
	} else {
		ZapLog.Debugf(template, args...)
	}
}

func (l *Logger) DebugKVs(ctx context.Context, msg string, kv ...any) {
	if ctx != nil {
		ZapLog.Debug(gcommon.GetTraceId(ctx) + gcommon.Kv2Str(msg, kv...))
	} else {
		ZapLog.Debug(gcommon.Kv2Str(msg, kv...))
	}
}

func (l *Logger) InfoF(ctx context.Context, template string, args ...interface{}) {
	if ctx != nil {
		ZapLog.Infof(gcommon.GetTraceId(ctx)+template, args...)
	} else {
		ZapLog.Infof(template, args...)
	}
}

func (l *Logger) InfoKVs(ctx context.Context, msg string, kv ...any) {

	if ctx != nil {
		ZapLog.Info(gcommon.GetTraceId(ctx) + gcommon.Kv2Str(msg, kv...))
	} else {
		ZapLog.Info(gcommon.Kv2Str(msg, kv...))
	}
}

func (l *Logger) WarnF(ctx context.Context, template string, args ...interface{}) {
	if ctx != nil {
		ZapLog.Warnf(gcommon.GetTraceId(ctx)+template, args...)
	} else {
		ZapLog.Warnf(template, args...)
	}
}

func (l *Logger) WarnKVs(ctx context.Context, msg string, kv ...any) {
	if ctx != nil {
		ZapLog.Warn(gcommon.GetTraceId(ctx) + gcommon.Kv2Str(msg, kv...))
	} else {
		ZapLog.Warn(gcommon.Kv2Str(msg, kv...))
	}
}

func (l *Logger) ErrorF(ctx context.Context, template string, args ...interface{}) {
	if ctx != nil {
		ZapLog.Errorf(gcommon.GetTraceId(ctx)+template, args...)
	} else {
		ZapLog.Errorf(template, args...)
	}
}

func (l *Logger) ErrorKVs(ctx context.Context, msg string, kv ...any) {
	if ctx != nil {
		ZapLog.Error(gcommon.GetTraceId(ctx) + gcommon.Kv2Str(msg, kv...))
	} else {
		ZapLog.Error(gcommon.Kv2Str(msg, kv...))
	}
}

func (l *Logger) PanicF(ctx context.Context, template string, args ...interface{}) {
	if ctx != nil {
		ZapLog.Panicf(gcommon.GetTraceId(ctx)+template, args...)
	} else {
		ZapLog.Panicf(template, args...)
	}
}

func (l *Logger) PanicKVs(ctx context.Context, msg string, kv ...any) {
	if ctx != nil {
		ZapLog.Panic(gcommon.GetTraceId(ctx) + gcommon.Kv2Str(msg, kv...))
	} else {
		ZapLog.Panic(gcommon.Kv2Str(msg, kv...))
	}
}

func (l *Logger) Printf(format string, v ...interface{}) {
	ZapLog.Infof(format, v...)
}

func (l *Logger) PrintKvs(msg string, kv ...any) {
	ZapLog.Info(gcommon.Kv2Str(msg, kv...))
}

func (l *Logger) Print(v ...interface{}) {
	ZapLog.Info(v...)
}

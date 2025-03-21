package tcp

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack2 "gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

func zapLog() *zap.SugaredLogger {

	writeSyncer := getLogWriter(fmt.Sprintf("%s/%s.log", "./log", "test"))

	encoder := DevEncoder()

	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)

	// zap.AddCaller()  添加将调用函数信息记录到日志中的功能。
	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
}

func getLogWriter(logPath string) zapcore.WriteSyncer {
	dir := path.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create log directory: %s", err))
	}
	return zapcore.AddSync(io.MultiWriter(&lumberjack2.Logger{
		Filename:   logPath,
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     1, //days
		LocalTime:  true,
		Compress:   true, // disabled by default
	}, os.Stdout))
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

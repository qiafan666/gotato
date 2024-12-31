package commons

import (
	"go.uber.org/zap/zapcore"
)

var ActiveRequests int64

const (
	Silent = iota - 2
	Debug
	Info
	Warn
	Error
)

var LogLevel = map[string]int{
	"silent": Silent,
	"debug":  Debug,
	"info":   Info,
	"warn":   Warn,
	"error":  Error,
}

var ZapLogLevel = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
}

const (
	Table = "table"
)

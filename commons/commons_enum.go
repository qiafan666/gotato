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

var DefaultLanguage = MsgLanguageEnglish

// msg language
const (
	MsgLanguageEnglish = "english"
	MsgLanguageChinese = "chinese"
)

const (
	Table  = "table"
	Layout = "2006-01-02 15:04:05"
)

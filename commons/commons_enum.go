package commons

import (
	"go.uber.org/zap/zapcore"
)

// ResponseCode define the error code
type ResponseCode int

var ActiveRequests int64

const (
	HttpNotFound   ResponseCode = -2
	UnKnowError    ResponseCode = -1
	OK             ResponseCode = 0
	ParameterError ResponseCode = 1
	ValidateError  ResponseCode = 2
	TokenError     ResponseCode = 3
	CheckAuthError ResponseCode = 4
)

const (
	CtxValueParameter = "parameter"
)

// CodeMsg global code and msg
var CodeMsg = map[string]map[ResponseCode]string{
	"english": {
		OK:             "suc",
		UnKnowError:    "unknown error",
		HttpNotFound:   "404",
		ParameterError: "parameter error",
		ValidateError:  "validate error",
		TokenError:     "Token error",
		CheckAuthError: "check auth error",
	},
	"chinese": {
		OK:             "成功",
		UnKnowError:    "未知错误",
		HttpNotFound:   "404",
		ParameterError: "参数错误",
		ValidateError:  "验证错误",
		TokenError:     "Token错误",
		CheckAuthError: "检查认证错误",
	},
}

// GetCodeAndMsg construct the code and msg
func GetCodeAndMsg(code ResponseCode, language string) string {
	if languageValue, ok := CodeMsg[language]; ok {
		if value, ok := languageValue[code]; ok {
			return value
		} else {
			return "{}"
		}
	} else {
		return "{}"
	}
}

// RegisterCodeAndMsg msg will be used as default msg, and you can change msg with function 'BuildFailedWithMsg' or 'BuildSuccessWithMsg' or 'response.WithMsg' for once.
func RegisterCodeAndMsg(language string, arr map[ResponseCode]string) {
	if len(arr) == 0 {
		return
	}
	for k, v := range arr {
		CodeMsg[language][k] = v
	}
}

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

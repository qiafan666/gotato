package gerr

import (
	"github.com/qiafan666/gotato/commons"
)

const (
	HttpNotFound   int = -2
	UnKnowError    int = -1
	OK             int = 0
	ParameterError int = 1
	ValidateError  int = 2
	TokenError     int = 3
	CheckAuthError int = 4
)

// CodeMsg global code and msg
var CodeMsg = map[string]map[int]string{
	commons.MsgLanguageEnglish: {
		OK:             "suc",
		UnKnowError:    "unknown error",
		HttpNotFound:   "404",
		ParameterError: "parameter error",
		ValidateError:  "validate error",
		TokenError:     "Token error",
		CheckAuthError: "check auth error",
	},
	commons.MsgLanguageChinese: {
		OK:             "成功",
		UnKnowError:    "未知错误",
		HttpNotFound:   "404",
		ParameterError: "参数错误",
		ValidateError:  "验证错误",
		TokenError:     "Token错误",
		CheckAuthError: "检查认证错误",
	},
}

// GetCodeAndMsg 获取code和msg
func GetCodeAndMsg(code int, language string) string {
	if languageValue, ok := CodeMsg[language]; ok {
		if value, ok := languageValue[code]; ok {
			return value
		} else {
			return ""
		}
	} else {
		return ""
	}
}

// RegisterCodeAndMsg 注册language的code和msg
func RegisterCodeAndMsg(language string, arr map[int]string) {
	if len(arr) == 0 {
		return
	}
	for k, v := range arr {
		CodeMsg[language][k] = v
	}
}

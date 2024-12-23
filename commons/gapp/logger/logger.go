package logger

import "github.com/qiafan666/gotato/commons/gface"

var DefaultLogger gface.Logger

func SetDefaultLogger(l gface.Logger) {
	DefaultLogger = l
}

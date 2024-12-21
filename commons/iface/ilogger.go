package iface

// Logger 定义一个通用的业务日志接口
type Logger interface {
	ErrorF(format string, args ...interface{})
	WarnF(format string, args ...interface{})
	DebugF(format string, args ...interface{})
	InfoF(format string, args ...interface{})
}

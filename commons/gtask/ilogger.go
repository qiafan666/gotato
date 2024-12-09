package gtask

// taskLogger 定义一个通用的日志接口
type taskLogger interface {
	TaskErrorF(format string, args ...interface{})
	TaskWarnF(format string, args ...interface{})
}

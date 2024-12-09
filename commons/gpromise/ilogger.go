package gpromise

// promiseLogger 定义一个通用的日志接口
type promiseLogger interface {
	PromiseErrorF(format string, args ...interface{})
	PromiseWarnF(format string, args ...interface{})
	PromiseDebugF(format string, args ...interface{})
	PromiseInfoF(format string, args ...interface{})
}

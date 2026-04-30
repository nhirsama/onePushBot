package platform

// Logger 定义平台层需要的最小日志能力。
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type discardLogger struct{}

func (discardLogger) Debug(string, ...any) {}
func (discardLogger) Info(string, ...any)  {}
func (discardLogger) Warn(string, ...any)  {}
func (discardLogger) Error(string, ...any) {}

// NewDiscardLogger 返回一个什么都不输出的日志实现。
func NewDiscardLogger() Logger {
	return discardLogger{}
}

package logger

import "io"

type Logger interface {
	Info(msg string, args ...interface{})
	Success(msg string, args ...interface{})
	Warning(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(args ...interface{})
	Print(args ...interface{})

	Writer() io.Writer // 新增：返回底层 io.Writer
}

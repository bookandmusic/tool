package logger

import (
	"fmt"
	"io"
	"time"

	"github.com/fatih/color"
)

// NewConsoleLogger 创建控制台 Logger
// 允许传入 io.Writer，例如 os.Stdout / 文件 / buffer
func NewConsoleLogger(out io.Writer, debug bool) Logger {
	console := &ConsoleLogger{
		debugEnabled: debug,
		out:          out,
	}
	return console
}

var (
	timeColor    = color.New(color.FgWhite).SprintFunc() // 时间统一颜色
	infoColor    = color.New(color.FgCyan).SprintFunc()
	successColor = color.New(color.FgGreen).SprintFunc()
	warnColor    = color.New(color.FgYellow).SprintFunc()
	errorColor   = color.New(color.FgRed).SprintFunc()
	debugColor   = color.New(color.FgHiBlack).SprintFunc()
)

// ConsoleLogger 控制台输出实现
type ConsoleLogger struct {
	debugEnabled bool
	out          io.Writer
}

// Writer 提供给外部库使用，例如 tablewriter
func (c *ConsoleLogger) Writer() io.Writer {
	return c.out
}

// containsPlaceholder 判断 msg 是否包含格式化占位符
func containsPlaceholder(msg string) bool {
	_, err := fmt.Fprintf(io.Discard, msg, nil)
	return err == nil
}

// formatMessage 增加时间戳、对齐级别标志和颜色
func (c *ConsoleLogger) formatMessage(level string, colorFunc func(a ...interface{}) string, msg string, args ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	var formattedMsg string
	if len(args) == 0 || !containsPlaceholder(msg) {
		formattedMsg = msg + fmt.Sprint(args...)
	} else {
		formattedMsg = fmt.Sprintf(msg, args...)
	}
	return fmt.Sprintf("%s %s %s", timeColor("["+timestamp+"]"), colorFunc(level), formattedMsg)
}

func (c *ConsoleLogger) Info(msg string, args ...interface{}) {
	fmt.Fprintln(c.out, c.formatMessage("INFO", infoColor, msg, args...))
}

func (c *ConsoleLogger) Success(msg string, args ...interface{}) {
	fmt.Fprintln(c.out, c.formatMessage("SUCCESS", successColor, msg, args...))
}

func (c *ConsoleLogger) Warning(msg string, args ...interface{}) {
	fmt.Fprintln(c.out, c.formatMessage("WARN", warnColor, msg, args...))
}

func (c *ConsoleLogger) Error(msg string, args ...interface{}) {
	fmt.Fprintln(c.out, c.formatMessage("ERROR", errorColor, msg, args...))
}

func (c *ConsoleLogger) Debug(args ...interface{}) {
	if c.debugEnabled {
		fmt.Fprint(c.out, debugColor(fmt.Sprint(args...)))
	}
}

func (c *ConsoleLogger) Print(args ...interface{}) {
	fmt.Fprint(c.out, args...)
}

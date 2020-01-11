package logger

import (
	"fmt"
	model "github.com/common/model"
	"os"
)

type ConsoleLogger struct {
	level int
}

func NewConsoleLogger(config model.LoggerConfig) (LogInterface, error) {
	logLevel := config.Level
	if logLevel == "" {
		return nil, fmt.Errorf("not found log_level")
	}

	return &ConsoleLogger{level: getLogLevel(logLevel)}, nil
}

func (c *ConsoleLogger) Init() {}

func (c *ConsoleLogger) SetLevel(level int) {
	if level < LogLevelDebug || level > LogLevelFatal {
		level = LogLevelDebug
	}

	c.level = level
}

func (c *ConsoleLogger) Debug(format string, args ...interface{}) {
	if c.level > LogLevelDebug {
		return
	}

	logData := writeLog(LogLevelDebug, format, args...)
	c.outputToConsole(logData)

}

func (c *ConsoleLogger) Trace(format string, args ...interface{}) {
	if c.level > LogLevelTrace {
		return
	}

	logData := writeLog(LogLevelTrace, format, args...)
	c.outputToConsole(logData)

}

func (c *ConsoleLogger) Info(format string, args ...interface{}) {
	if c.level > LogLevelInfo {
		return
	}

	logData := writeLog(LogLevelInfo, format, args...)
	c.outputToConsole(logData)

}

func (c *ConsoleLogger) Warn(format string, args ...interface{}) {
	if c.level > LogLevelWarn {
		return
	}

	logData := writeLog(LogLevelWarn, format, args...)
	c.outputToConsole(logData)

}

func (c *ConsoleLogger) Error(format string, args ...interface{}) {
	if c.level > LogLevelError {
		return
	}

	logData := writeLog(LogLevelError, format, args...)
	c.outputToConsole(logData)

}

func (c *ConsoleLogger) Fatal(format string, args ...interface{}) {
	if c.level > LogLevelFatal {
		return
	}

	logData := writeLog(LogLevelFatal, format, args...)
	c.outputToConsole(logData)
}

func (c *ConsoleLogger) Close() {}

func (c *ConsoleLogger) outputToConsole(logData *LogData) {
	fmt.Fprintf(os.Stdout, "%s %s (%s:%s:%d) %s\n", logData.TimeStr,
		logData.LevelStr, logData.Filename, logData.FuncName, logData.LineNo, logData.Message)
}

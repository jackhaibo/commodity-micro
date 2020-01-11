package logger

import (
	"fmt"
	model "github.com/common/model"
)

var log LogInterface
var preLog LogInterface

/*
file, "初始化文件日志实例"
*/
func InitLogger(config model.LoggerConfig) (err error) {
	if log != nil {
		preLog = log
		defer preLog.Close()
	}
	switch config.Type {
	case "file":
		log, err = NewFileLogger(config)
	case "console":
		log, err = NewConsoleLogger(config)
	default:
		err = fmt.Errorf("unsupport logger type:%s", config.Type)
	}

	return
}

func Debug(format string, args ...interface{}) {
	log.Debug(format, args...)
}

func Trace(format string, args ...interface{}) {
	log.Trace(format, args...)
}

func Info(format string, args ...interface{}) {
	log.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	log.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	log.Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	log.Fatal(format, args...)
}

package logger

import (
	"fmt"
	"path"
	"runtime"
	"time"
)

type LogData struct {
	Message  string
	TimeStr  string
	LevelStr string
	Filename string
	FuncName string
	LineNo   int
}

func GetLineInfo() (fileName string, funcName string, lineNo int) {
	pc, file, line, ok := runtime.Caller(4)
	if ok {
		fileName = file
		funcName = runtime.FuncForPC(pc).Name()
		lineNo = line
	}
	return
}

func writeLog(level int, format string, args ...interface{}) *LogData {
	now := time.Now()
	nowStr := now.Format("2006-01-02 15:04:05.999")
	levelStr := getLevelText(level)

	fileName, funcName, lineNo := GetLineInfo()
	fileName = path.Base(fileName)
	funcName = path.Base(funcName)
	msg := fmt.Sprintf(format, args...)

	logData := &LogData{
		Message:  msg,
		TimeStr:  nowStr,
		LevelStr: levelStr,
		Filename: fileName,
		FuncName: funcName,
		LineNo:   lineNo,
	}

	return logData
}

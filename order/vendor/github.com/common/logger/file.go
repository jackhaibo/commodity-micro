package logger

import (
	"fmt"
	"github.com/common/model"
	"os"
	"strconv"
	"time"
)

type FileLogger struct {
	level         int
	logPath       string
	logName       string
	file          *os.File
	LogDataChan   chan *LogData
	logSplitType  int
	logSplitSize  int64
	lastSplitHour int
}

func NewFileLogger(config model.LoggerConfig) (LogInterface, error) {
	logPath := config.Path
	if logPath == "" {
		return nil, fmt.Errorf("not found log_path")
	}

	logName := config.Name
	if logName == "" {
		return nil, fmt.Errorf("not found log_name ")
	}

	logLevel := config.Level
	if logLevel == "" {
		return nil, fmt.Errorf("not found log_level ")
	}

	logChanSize := config.ChanSize
	if logChanSize == "" {
		logChanSize = LogChanSize
	}

	chanSize, err := strconv.Atoi(logChanSize)
	if err != nil {
		chanSize = ChanSize
	}

	var logSplitType int
	var logSplitSize int64
	logSplitStr := config.SplitType
	if logSplitStr == "" {
		logSplitType = LogSplitTypeHour
	} else {
		if logSplitStr == "size" {
			logSplitSizeStr := config.SplitSize
			if logSplitSizeStr == "" {
				logSplitSizeStr = LogSplitSizeStr
			}

			logSplitSize, err = strconv.ParseInt(logSplitSizeStr, 10, 64)
			if err != nil {
				logSplitSize = LogSplitSize
			}
			logSplitType = LogSplitTypeSize
		} else {
			logSplitType = LogSplitTypeHour
		}
	}

	level := getLogLevel(logLevel)
	log = &FileLogger{
		level:         level,
		logPath:       logPath,
		logName:       logName,
		LogDataChan:   make(chan *LogData, chanSize),
		logSplitSize:  logSplitSize,
		logSplitType:  logSplitType,
		lastSplitHour: time.Now().Hour(),
	}

	return log, nil
}

func (f *FileLogger) Init() {
	filename := fmt.Sprintf("%s/%s.log", f.logPath, f.logName)
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		panic(fmt.Sprintf("open faile %s failed, err:%v", filename, err))
	}

	f.file = file

	go f.writeLogBackground()
}

func (f *FileLogger) splitFileHour() {
	now := time.Now()
	hour := now.Hour()
	if hour == f.lastSplitHour {
		return
	}

	f.lastSplitHour = hour
	var backupFilename string
	var filename string

	backupFilename = fmt.Sprintf("%s/%s.log_%04d%02d%02d%02d",
		f.logPath, f.logName, now.Year(), now.Month(), now.Day(), f.lastSplitHour)
	filename = fmt.Sprintf("%s/%s.log", f.logPath, f.logName)

	f.file.Close()

	os.Rename(filename, backupFilename)

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return
	}

	f.file = file
}

func (f *FileLogger) splitFileSize() {
	file := f.file

	statInfo, err := file.Stat()
	if err != nil {
		return
	}

	fileSize := statInfo.Size()
	if fileSize <= f.logSplitSize {
		return
	}

	var backupFilename string
	var filename string

	now := time.Now()

	backupFilename = fmt.Sprintf("%s/%s.log_%04d%02d%02d%02d%02d%02d",
		f.logPath, f.logName, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
	filename = fmt.Sprintf("%s/%s.log", f.logPath, f.logName)

	file.Close()
	os.Rename(filename, backupFilename)

	file, err = os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return
	}

	f.file = file
}

func (f *FileLogger) checkSplitFileType() {
	if f.logSplitType == LogSplitTypeHour {
		f.splitFileHour()
		return
	}

	f.splitFileSize()
}

func (f *FileLogger) writeLogBackground() {
	for logData := range f.LogDataChan {
		var file = f.file

		f.checkSplitFileType()

		fmt.Fprintf(file, "%s %s (%s:%s:%d) %s\n", logData.TimeStr,
			logData.LevelStr, logData.Filename, logData.FuncName, logData.LineNo, logData.Message)
	}
}

func (f *FileLogger) SetLevel(level int) {
	if level < LogLevelDebug || level > LogLevelFatal {
		level = LogLevelDebug
	}
	f.level = level
}

func (f *FileLogger) Debug(format string, args ...interface{}) {
	if f.level > LogLevelDebug {
		return
	}

	logData := writeLog(LogLevelDebug, format, args...)
	select {
	case f.LogDataChan <- logData:
	default:
	}
}

func (f *FileLogger) Trace(format string, args ...interface{}) {
	if f.level > LogLevelTrace {
		return
	}
	logData := writeLog(LogLevelTrace, format, args...)
	select {
	case f.LogDataChan <- logData:
	default:
	}
}

func (f *FileLogger) Info(format string, args ...interface{}) {
	if f.level > LogLevelInfo {
		return
	}
	logData := writeLog(LogLevelInfo, format, args...)
	select {
	case f.LogDataChan <- logData:
	default:
	}
}

func (f *FileLogger) Warn(format string, args ...interface{}) {
	if f.level > LogLevelWarn {
		return
	}

	logData := writeLog(LogLevelWarn, format, args...)
	select {
	case f.LogDataChan <- logData:
	default:
	}
}

func (f *FileLogger) Error(format string, args ...interface{}) {
	if f.level > LogLevelError {
		return
	}

	logData := writeLog(LogLevelError, format, args...)
	select {
	case f.LogDataChan <- logData:
	default:
	}
}

func (f *FileLogger) Fatal(format string, args ...interface{}) {
	if f.level > LogLevelFatal {
		return
	}

	logData := writeLog(LogLevelFatal, format, args...)
	select {
	case f.LogDataChan <- logData:
	default:
	}
}

func (f *FileLogger) Close() {
	f.file.Close()
}

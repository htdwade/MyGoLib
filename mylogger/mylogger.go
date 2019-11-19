package mylogger

import (
	"errors"
	"fmt"
	"path"
	"runtime"
	"strings"
)

// 自定义一个日志库

type LogLevel uint16

// Logger 接口
type Logger interface {
	Debug(format string, a ...interface{})
	Info(format string, a ...interface{})
	Warning(format string, a ...interface{})
	Error(format string, a ...interface{})
	Fatal(format string, a ...interface{})
}

const (
	UNKNOWN LogLevel = iota
	DEBUG
	INFO
	WARNING
	ERROR
	FATAL
)

func parseLogLevel(logLevelStr string) (LogLevel, error) {
	s := strings.ToLower(logLevelStr)
	switch s {
	case "debug":
		return DEBUG, nil
	case "info":
		return INFO, nil
	case "warning":
		return WARNING, nil
	case "error":
		return ERROR, nil
	case "fatal":
		return FATAL, nil
	default:
		err := errors.New("无效的日志级别")
		return UNKNOWN, err
	}
}

func getLogString(logLevel LogLevel) string {
	switch logLevel {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	}
	return "UNKNOWN"
}

func getInfo(skip int) (fileName, funcName string, line int) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		fmt.Printf("runtime.Caller() failed\n")
		return
	}
	fileName = path.Base(file)
	funcName = runtime.FuncForPC(pc).Name()
	funcName = strings.Split(funcName, ".")[1]
	return
}

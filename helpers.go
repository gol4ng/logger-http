package logger_http

import (
	"runtime"
	"strconv"
)

func MessageWithFileLine(message string, skip int) string {
	_, file, line, ok := runtime.Caller(1 + skip)
	if ok {
		message += " a" + file + ":" + strconv.Itoa(line)
	}
	return message
}

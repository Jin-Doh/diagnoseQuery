package log

import (
	"fmt"
	"time"
)

const (
	LogLevelDebug   = "DEBUG"
	LogLevelInfo    = "INFO"
	LogLevelWarning = "WARNING"
	LogLevelError   = "ERROR"
	LogLevelFatal   = "FATAL"
)

// 통일된 로그 포맷 함수
func Message(level, format string, v ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, v...)
	fmt.Printf("[%s] [%s] %s\n", timestamp, level, msg)
	if level == LogLevelFatal {
		panic(msg)
	}
}

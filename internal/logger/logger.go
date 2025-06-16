package logger

import (
	"io"
	"log"
	"os"
)

var debugLogger *log.Logger

func SetDebugEnabled(enabled bool) {
	if enabled {
		debugLogger = log.New(os.Stderr, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
		debugLogger.Println("Debug logging enabled")
	} else {
		debugLogger = log.New(io.Discard, "", 0)
	}
}

func Debug(format string, args ...interface{}) {
	if debugLogger != nil {
		debugLogger.Printf(format, args...)
	}
}

func IsDebugEnabled() bool {
	return debugLogger != nil && debugLogger.Writer() != io.Discard
}

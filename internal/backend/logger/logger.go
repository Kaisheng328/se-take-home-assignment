package logger

import (
	"fmt"
	"io"
	"time"
)

// Logger provides timestamped logging for the order management system.
type Logger struct {
	writer io.Writer
}

// New creates a new Logger that writes to the given writer.
func New(w io.Writer) *Logger {
	return &Logger{writer: w}
}

// Log writes a timestamped message in [HH:MM:SS] format.
func (l *Logger) Log(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(l.writer, "[%s] %s\n", timestamp, message)
}

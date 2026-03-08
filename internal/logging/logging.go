package logging

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// Level represents a logging severity level.
type Level int

const (
	LevelError Level = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

var (
	mu       sync.Mutex
	level    = LevelWarn
	output   io.Writer = os.Stderr
)

// SetLevel sets the global log level.
func SetLevel(l Level) {
	mu.Lock()
	defer mu.Unlock()
	level = l
}

// SetOutput sets the output writer (default: os.Stderr). Mainly for testing.
func SetOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	output = w
}

// Errorf logs a message at Error level.
func Errorf(format string, args ...interface{}) {
	log(LevelError, "ERROR", format, args...)
}

// Warn logs a message at Warn level.
func Warn(format string, args ...interface{}) {
	log(LevelWarn, "WARN", format, args...)
}

// Info logs a message at Info level.
func Info(format string, args ...interface{}) {
	log(LevelInfo, "INFO", format, args...)
}

// Debug logs a message at Debug level.
func Debug(format string, args ...interface{}) {
	log(LevelDebug, "DEBUG", format, args...)
}

func log(msgLevel Level, label, format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if msgLevel > level {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(output, "[%s] %s\n", label, msg)
}

// Writer returns an io.Writer that writes each line as a Debug-level log message.
// This is intended for integration with ratelimit.SetLogWriter().
func Writer() io.Writer {
	return &logWriter{}
}

type logWriter struct{}

func (w *logWriter) Write(p []byte) (n int, err error) {
	// Strip trailing newline since log() adds one.
	msg := string(p)
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}
	Debug("%s", msg)
	return len(p), nil
}

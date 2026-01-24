package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Level represents log severity
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Field represents a structured log field
type Field struct {
	Key   string
	Value any
}

// F creates a new log field
func F(key string, value any) Field {
	return Field{Key: key, Value: value}
}

// Logger defines the logging interface
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, err error, fields ...Field)
	WithFields(fields ...Field) Logger
}

// SimpleLogger is a basic implementation of Logger
type SimpleLogger struct {
	mu       sync.Mutex
	out      io.Writer
	level    Level
	fields   []Field
}

// Config holds logger configuration
type Config struct {
	Level  Level
	Output io.Writer
}

// New creates a new SimpleLogger
func New(cfg Config) *SimpleLogger {
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}
	return &SimpleLogger{
		out:    cfg.Output,
		level:  cfg.Level,
		fields: nil,
	}
}

// NewDefault creates a logger with default settings
func NewDefault(debug bool) *SimpleLogger {
	level := LevelInfo
	if debug {
		level = LevelDebug
	}
	return New(Config{
		Level:  level,
		Output: os.Stdout,
	})
}

func (l *SimpleLogger) log(level Level, msg string, err error, fields []Field) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format(time.RFC3339)

	// Build log line
	line := fmt.Sprintf("%s [%s] %s", timestamp, level.String(), msg)

	// Add base fields
	allFields := append(l.fields, fields...)

	// Add error if present
	if err != nil {
		allFields = append(allFields, F("error", err.Error()))
	}

	// Add fields
	for _, f := range allFields {
		line += fmt.Sprintf(" %s=%v", f.Key, f.Value)
	}

	fmt.Fprintln(l.out, line)
}

// Debug logs a debug message
func (l *SimpleLogger) Debug(msg string, fields ...Field) {
	l.log(LevelDebug, msg, nil, fields)
}

// Info logs an info message
func (l *SimpleLogger) Info(msg string, fields ...Field) {
	l.log(LevelInfo, msg, nil, fields)
}

// Warn logs a warning message
func (l *SimpleLogger) Warn(msg string, fields ...Field) {
	l.log(LevelWarn, msg, nil, fields)
}

// Error logs an error message
func (l *SimpleLogger) Error(msg string, err error, fields ...Field) {
	l.log(LevelError, msg, err, fields)
}

// WithFields returns a new logger with additional fields
func (l *SimpleLogger) WithFields(fields ...Field) Logger {
	newFields := make([]Field, len(l.fields)+len(fields))
	copy(newFields, l.fields)
	copy(newFields[len(l.fields):], fields)

	return &SimpleLogger{
		out:    l.out,
		level:  l.level,
		fields: newFields,
	}
}

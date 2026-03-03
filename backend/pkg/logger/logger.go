package logger

import (
	"io"
	"log/slog"
	"os"
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

// SimpleLogger is a slog-backed implementation of Logger.
type SimpleLogger struct {
	out    io.Writer
	level  Level
	fields []slog.Attr
	base   *slog.Logger
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
	handler := slog.NewTextHandler(cfg.Output, &slog.HandlerOptions{
		Level: levelToSlog(cfg.Level),
	})
	base := slog.New(handler)
	return &SimpleLogger{
		out:    cfg.Output,
		level:  cfg.Level,
		fields: nil,
		base:   base,
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
	if l.base == nil {
		return
	}
	attrs := make([]slog.Attr, 0, len(l.fields)+len(fields)+1)
	attrs = append(attrs, l.fields...)
	for _, f := range fields {
		attrs = append(attrs, slog.Any(f.Key, f.Value))
	}
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
	}
	l.base.LogAttrs(nil, levelToSlog(level), msg, attrs...)
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
	newAttrs := make([]slog.Attr, len(l.fields), len(l.fields)+len(fields))
	copy(newAttrs, l.fields)
	for _, f := range fields {
		newAttrs = append(newAttrs, slog.Any(f.Key, f.Value))
	}

	return &SimpleLogger{
		out:    l.out,
		level:  l.level,
		fields: newAttrs,
		base:   l.base,
	}
}

func levelToSlog(level Level) slog.Level {
	switch level {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

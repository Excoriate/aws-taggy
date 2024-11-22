package o11y

import (
	"io"
	"log/slog"
	"os"
)

// LogLevel represents the logging verbosity
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// Logger provides a structured logging interface
type Logger struct {
	logger *slog.Logger
	level  LogLevel
}

// NewLogger creates a new logger with specified options
func NewLogger(output io.Writer, level LogLevel) *Logger {
	var slogLevel slog.Level

	switch level {
	case LogLevelDebug:
		slogLevel = slog.LevelDebug
	case LogLevelInfo:
		slogLevel = slog.LevelInfo
	case LogLevelWarn:
		slogLevel = slog.LevelWarn
	case LogLevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	handler := slog.NewTextHandler(output, &slog.HandlerOptions{
		Level: slogLevel,
	})

	return &Logger{
		logger: slog.New(handler),
		level:  level,
	}
}

// DefaultLogger creates a logger with default settings
func DefaultLogger() *Logger {
	return NewLogger(os.Stdout, LogLevelInfo)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

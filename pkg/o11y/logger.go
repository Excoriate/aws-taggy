package o11y

import (
	"io"
	"os"

	"github.com/charmbracelet/log"
)

// LogLevel represents the logging verbosity
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// Logger provides a structured logging interface with emojis
type Logger struct {
	logger *log.Logger
	level  LogLevel
}

// LoggerInterface defines the contract for logging methods
type LoggerInterface interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// Ensure Logger implements LoggerInterface
var _ LoggerInterface = (*Logger)(nil)

// NewLogger creates a new logger with specified options and emojis
func NewLogger(output io.Writer, level LogLevel) *Logger {
	if output == nil {
		output = os.Stdout
	}

	// Create a new Charmbracelet logger
	charmLogger := log.New(output)

	// Set log level with emoji support
	switch level {
	case LogLevelDebug:
		charmLogger.SetLevel(log.DebugLevel)
		charmLogger.SetReportCaller(true)
	case LogLevelInfo:
		charmLogger.SetLevel(log.InfoLevel)
	case LogLevelWarn:
		charmLogger.SetLevel(log.WarnLevel)
	case LogLevelError:
		charmLogger.SetLevel(log.ErrorLevel)
	default:
		charmLogger.SetLevel(log.InfoLevel)
	}

	// Configure emoji and styling
	charmLogger.SetFormatter(log.TextFormatter)
	charmLogger.SetReportTimestamp(true)

	return &Logger{
		logger: charmLogger,
		level:  level,
	}
}

// DefaultLogger creates a logger with default settings and emojis
func DefaultLogger() *Logger {
	return NewLogger(os.Stdout, LogLevelInfo)
}

// Debug logs a debug message with 🐞 emoji
func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug("🐞 "+msg, args...)
}

// Info logs an info message with 📝 emoji
func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info("ℹ️ "+msg, args...)
}

// Warn logs a warning message with ⚠️ emoji
func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn("🔔 "+msg, args...)
}

// Error logs an error message with 🚨 emoji
func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error("🚨 "+msg, args...)
}

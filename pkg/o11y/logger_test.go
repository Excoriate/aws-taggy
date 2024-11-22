package o11y

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	testCases := []struct {
		name          string
		logLevel      LogLevel
		logFunc       func(*Logger, string, ...any)
		expectedLevel string
	}{
		{
			"Debug Level", 
			LogLevelDebug, 
			func(l *Logger, msg string, args ...any) { l.Debug(msg, args...) }, 
			"level=DEBUG",
		},
		{
			"Info Level", 
			LogLevelInfo, 
			func(l *Logger, msg string, args ...any) { l.Info(msg, args...) }, 
			"level=INFO",
		},
		{
			"Warn Level", 
			LogLevelWarn, 
			func(l *Logger, msg string, args ...any) { l.Warn(msg, args...) }, 
			"level=WARN",
		},
		{
			"Error Level", 
			LogLevelError, 
			func(l *Logger, msg string, args ...any) { l.Error(msg, args...) }, 
			"level=ERROR",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(&buf, tc.logLevel)

			tc.logFunc(logger, "Test log message")
			logOutput := buf.String()

			assert.True(t, 
				strings.Contains(logOutput, tc.expectedLevel), 
				"Expected log level %s not found in output", 
				tc.expectedLevel,
			)
			assert.True(t, 
				strings.Contains(logOutput, "Test log message"), 
				"Log message not found in output",
			)
		})
	}
}

func TestDefaultLogger(t *testing.T) {
	defaultLogger := DefaultLogger()
	assert.NotNil(t, defaultLogger)
}

func TestLoggerMethods(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, LogLevelDebug)

	testCases := []struct {
		name     string
		logFunc  func(string, ...any)
		message  string
		expected string
	}{
		{"Debug", logger.Debug, "Debug message", "level=DEBUG"},
		{"Info", logger.Info, "Info message", "level=INFO"},
		{"Warn", logger.Warn, "Warn message", "level=WARN"},
		{"Error", logger.Error, "Error message", "level=ERROR"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			tc.logFunc(tc.message)
			logOutput := buf.String()

			assert.True(t, 
				strings.Contains(logOutput, tc.expected), 
				"Expected log level %s not found in output", 
				tc.expected,
			)
			assert.True(t, 
				strings.Contains(logOutput, tc.message), 
				"Log message not found in output",
			)
		})
	}
}

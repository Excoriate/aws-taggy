package o11y

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		logFunc  func(*Logger, string)
		expected string
	}{
		{
			name:  "Debug Level",
			level: LogLevelDebug,
			logFunc: func(l *Logger, msg string) {
				l.Debug(msg)
			},
			expected: "🐞",
		},
		{
			name:  "Info Level",
			level: LogLevelInfo,
			logFunc: func(l *Logger, msg string) {
				l.Info(msg)
			},
			expected: "ℹ️",
		},
		{
			name:  "Warn Level",
			level: LogLevelWarn,
			logFunc: func(l *Logger, msg string) {
				l.Warn(msg)
			},
			expected: "🔔",
		},
		{
			name:  "Error Level",
			level: LogLevelError,
			logFunc: func(l *Logger, msg string) {
				l.Error(msg)
			},
			expected: "🚨",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(&buf, tt.level)

			tt.logFunc(logger, "test message")
			output := buf.String()
			assert.True(t, strings.Contains(output, tt.expected),
				"Expected emoji %s not found in output: %s", tt.expected, output)
			assert.True(t, strings.Contains(output, "test message"),
				"Expected message not found in output: %s", output)
		})
	}
}

func TestDefaultLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, LogLevelInfo)

	logger.Info("test message")
	output := buf.String()
	assert.True(t, strings.Contains(output, "ℹ️"),
		"Expected info emoji not found in output: %s", output)
	assert.True(t, strings.Contains(output, "test message"),
		"Expected message not found in output: %s", output)
}

func TestLoggerMethods(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		logFunc  func(*Logger, string)
		expected string
	}{
		{
			name:  "Debug",
			level: LogLevelDebug,
			logFunc: func(l *Logger, msg string) {
				l.Debug(msg)
			},
			expected: "🐞",
		},
		{
			name:  "Info",
			level: LogLevelInfo,
			logFunc: func(l *Logger, msg string) {
				l.Info(msg)
			},
			expected: "ℹ️",
		},
		{
			name:  "Warn",
			level: LogLevelWarn,
			logFunc: func(l *Logger, msg string) {
				l.Warn(msg)
			},
			expected: "🔔",
		},
		{
			name:  "Error",
			level: LogLevelError,
			logFunc: func(l *Logger, msg string) {
				l.Error(msg)
			},
			expected: "🚨",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(&buf, tt.level)

			tt.logFunc(logger, "test message")
			output := buf.String()
			assert.True(t, strings.Contains(output, tt.expected),
				"Expected emoji %s not found in output: %s", tt.expected, output)
			assert.True(t, strings.Contains(output, "test message"),
				"Expected message not found in output: %s", output)
		})
	}
}

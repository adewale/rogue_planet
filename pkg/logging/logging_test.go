package logging

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestStandardLoggerSetLevel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		level string
		want  Level
	}{
		{"error level", "error", LevelError},
		{"warn level", "warn", LevelWarn},
		{"warning level", "warning", LevelWarn},
		{"info level", "info", LevelInfo},
		{"debug level", "debug", LevelDebug},
		{"unknown defaults to info", "unknown", LevelInfo},
		{"empty defaults to info", "", LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewWithLevel(LevelInfo)
			l.SetLevel(tt.level)
			if l.level != tt.want {
				t.Errorf("SetLevel(%q) level = %d, want %d", tt.level, l.level, tt.want)
			}
		})
	}
}

func TestStandardLoggerLevelFiltering(t *testing.T) {
	tests := []struct {
		name        string
		logLevel    string
		logFunc     func(*StandardLogger)
		shouldLog   bool
		containsMsg string
	}{
		{
			name:        "error logs at error level",
			logLevel:    "error",
			logFunc:     func(l *StandardLogger) { l.Error("test error") },
			shouldLog:   true,
			containsMsg: "ERROR: test error",
		},
		{
			name:        "warn does not log at error level",
			logLevel:    "error",
			logFunc:     func(l *StandardLogger) { l.Warn("test warn") },
			shouldLog:   false,
			containsMsg: "WARN: test warn",
		},
		{
			name:        "info does not log at error level",
			logLevel:    "error",
			logFunc:     func(l *StandardLogger) { l.Info("test info") },
			shouldLog:   false,
			containsMsg: "INFO: test info",
		},
		{
			name:        "debug does not log at error level",
			logLevel:    "error",
			logFunc:     func(l *StandardLogger) { l.Debug("test debug") },
			shouldLog:   false,
			containsMsg: "DEBUG: test debug",
		},
		{
			name:        "error logs at warn level",
			logLevel:    "warn",
			logFunc:     func(l *StandardLogger) { l.Error("test error") },
			shouldLog:   true,
			containsMsg: "ERROR: test error",
		},
		{
			name:        "warn logs at warn level",
			logLevel:    "warn",
			logFunc:     func(l *StandardLogger) { l.Warn("test warn") },
			shouldLog:   true,
			containsMsg: "WARN: test warn",
		},
		{
			name:        "info does not log at warn level",
			logLevel:    "warn",
			logFunc:     func(l *StandardLogger) { l.Info("test info") },
			shouldLog:   false,
			containsMsg: "INFO: test info",
		},
		{
			name:        "error logs at info level",
			logLevel:    "info",
			logFunc:     func(l *StandardLogger) { l.Error("test error") },
			shouldLog:   true,
			containsMsg: "ERROR: test error",
		},
		{
			name:        "warn logs at info level",
			logLevel:    "info",
			logFunc:     func(l *StandardLogger) { l.Warn("test warn") },
			shouldLog:   true,
			containsMsg: "WARN: test warn",
		},
		{
			name:        "info logs at info level",
			logLevel:    "info",
			logFunc:     func(l *StandardLogger) { l.Info("test info") },
			shouldLog:   true,
			containsMsg: "INFO: test info",
		},
		{
			name:        "debug does not log at info level",
			logLevel:    "info",
			logFunc:     func(l *StandardLogger) { l.Debug("test debug") },
			shouldLog:   false,
			containsMsg: "DEBUG: test debug",
		},
		{
			name:        "error logs at debug level",
			logLevel:    "debug",
			logFunc:     func(l *StandardLogger) { l.Error("test error") },
			shouldLog:   true,
			containsMsg: "ERROR: test error",
		},
		{
			name:        "debug logs at debug level",
			logLevel:    "debug",
			logFunc:     func(l *StandardLogger) { l.Debug("test debug") },
			shouldLog:   true,
			containsMsg: "DEBUG: test debug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var buf bytes.Buffer
			log.SetOutput(&buf)
			log.SetFlags(0) // Remove timestamp for easier testing
			defer func() {
				log.SetOutput(nil)
				log.SetFlags(log.LstdFlags)
			}()

			// Create logger and set level
			l := New(tt.logLevel)

			// Call the log function
			tt.logFunc(l)

			// Check output
			output := buf.String()
			if tt.shouldLog {
				if !strings.Contains(output, tt.containsMsg) {
					t.Errorf("Expected log to contain %q, got %q", tt.containsMsg, output)
				}
			} else {
				if output != "" {
					t.Errorf("Expected no log output at level %q, got %q", tt.logLevel, output)
				}
			}
		})
	}
}

func TestStandardLoggerFormatting(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetFlags(0)
	defer func() {
		log.SetOutput(nil)
		log.SetFlags(log.LstdFlags)
	}()

	l := NewWithLevel(LevelDebug)

	l.Info("test %s %d", "message", 42)

	output := buf.String()
	expected := "INFO: test message 42"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected log to contain %q, got %q", expected, output)
	}
}

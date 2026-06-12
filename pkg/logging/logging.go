// Package logging provides a simple leveled logging interface and implementation.
package logging

import "log"

// Logger is the interface for structured logging with levels.
// Implementations should format messages consistently and respect the configured level.
type Logger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)
}

// Level represents the logging level
type Level int

const (
	LevelError Level = 0
	LevelWarn  Level = 1
	LevelInfo  Level = 2
	LevelDebug Level = 3
)

// ParseLevel converts a level string to a Level constant.
// Valid levels: "error", "warn"/"warning", "info", "debug".
// Returns LevelInfo for unrecognized values.
func ParseLevel(levelStr string) Level {
	switch levelStr {
	case "error":
		return LevelError
	case "warn", "warning":
		return LevelWarn
	case "info":
		return LevelInfo
	case "debug":
		return LevelDebug
	default:
		return LevelInfo
	}
}

// StandardLogger wraps Go's standard logger with level support.
// Messages are prefixed with their level (ERROR, WARN, INFO, DEBUG).
type StandardLogger struct {
	level Level
}

// New creates a new StandardLogger with the specified level.
// Valid levels: "error", "warn"/"warning", "info", "debug".
// Defaults to info if level is unrecognized.
func New(levelStr string) *StandardLogger {
	return &StandardLogger{level: ParseLevel(levelStr)}
}

// NewWithLevel creates a new StandardLogger with the specified level constant.
func NewWithLevel(level Level) *StandardLogger {
	return &StandardLogger{level: level}
}

// SetLevel changes the logger's level at runtime.
func (l *StandardLogger) SetLevel(levelStr string) {
	l.level = ParseLevel(levelStr)
}

// Error logs an error message.
func (l *StandardLogger) Error(format string, args ...any) {
	if l.level >= LevelError {
		log.Printf("ERROR: "+format, args...)
	}
}

// Warn logs a warning message.
func (l *StandardLogger) Warn(format string, args ...any) {
	if l.level >= LevelWarn {
		log.Printf("WARN: "+format, args...)
	}
}

// Info logs an info message.
func (l *StandardLogger) Info(format string, args ...any) {
	if l.level >= LevelInfo {
		log.Printf("INFO: "+format, args...)
	}
}

// Debug logs a debug message.
func (l *StandardLogger) Debug(format string, args ...any) {
	if l.level >= LevelDebug {
		log.Printf("DEBUG: "+format, args...)
	}
}

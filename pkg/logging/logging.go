// Package logging provides a simple leveled logging interface and implementation.
package logging

import "log"

// Logger is the interface for structured logging with levels.
// Implementations should format messages consistently and respect the configured level.
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
}

// Level represents the logging level
type Level int

const (
	LevelError Level = 0
	LevelWarn  Level = 1
	LevelInfo  Level = 2
	LevelDebug Level = 3
)

// StandardLogger wraps Go's standard logger with level support.
// Messages are prefixed with their level (ERROR, WARN, INFO, DEBUG).
type StandardLogger struct {
	level Level
}

// New creates a new StandardLogger with the specified level.
// Valid levels: "error", "warn"/"warning", "info", "debug".
// Defaults to info if level is unrecognized.
func New(levelStr string) *StandardLogger {
	var level Level
	switch levelStr {
	case "error":
		level = LevelError
	case "warn", "warning":
		level = LevelWarn
	case "info":
		level = LevelInfo
	case "debug":
		level = LevelDebug
	default:
		level = LevelInfo
	}
	return &StandardLogger{level: level}
}

// NewWithLevel creates a new StandardLogger with the specified level constant.
func NewWithLevel(level Level) *StandardLogger {
	return &StandardLogger{level: level}
}

// SetLevel changes the logger's level at runtime.
func (l *StandardLogger) SetLevel(levelStr string) {
	switch levelStr {
	case "error":
		l.level = LevelError
	case "warn", "warning":
		l.level = LevelWarn
	case "info":
		l.level = LevelInfo
	case "debug":
		l.level = LevelDebug
	default:
		l.level = LevelInfo
	}
}

// Error logs an error message.
func (l *StandardLogger) Error(format string, args ...interface{}) {
	if l.level >= LevelError {
		log.Printf("ERROR: "+format, args...)
	}
}

// Warn logs a warning message.
func (l *StandardLogger) Warn(format string, args ...interface{}) {
	if l.level >= LevelWarn {
		log.Printf("WARN: "+format, args...)
	}
}

// Info logs an info message.
func (l *StandardLogger) Info(format string, args ...interface{}) {
	if l.level >= LevelInfo {
		log.Printf("INFO: "+format, args...)
	}
}

// Debug logs a debug message.
func (l *StandardLogger) Debug(format string, args ...interface{}) {
	if l.level >= LevelDebug {
		log.Printf("DEBUG: "+format, args...)
	}
}

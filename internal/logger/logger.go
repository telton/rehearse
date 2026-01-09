// Package logger provides centralized logging configuration for rehearse.
package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

var globalLogger *slog.Logger

// Level represents log levels
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Config holds logger configuration
type Config struct {
	Level  Level
	Format string // "text" or "json"
	Output io.Writer
}

// DefaultConfig returns default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:  LevelInfo,
		Format: "text",
		Output: os.Stdout,
	}
}

// Setup initializes the global logger with the given configuration
func Setup(cfg *Config) {
	level := parseLevel(cfg.Level)

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(cfg.Output, opts)
	} else {
		handler = slog.NewTextHandler(cfg.Output, opts)
	}

	globalLogger = slog.New(handler)
	slog.SetDefault(globalLogger)
}

// parseLevel converts string level to slog.Level
func parseLevel(level Level) slog.Level {
	switch strings.ToLower(string(level)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// ParseLevelFromString parses log level from string (for env var and flags)
func ParseLevelFromString(s string) Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// Get returns the global logger instance
func Get() *slog.Logger {
	if globalLogger == nil {
		Setup(DefaultConfig())
	}
	return globalLogger
}

// Debug logs at debug level
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

// Info logs at info level
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

// Warn logs at warn level
func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

// Error logs at error level
func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

// With returns a new logger with the given attributes
func With(args ...any) *slog.Logger {
	return Get().With(args...)
}

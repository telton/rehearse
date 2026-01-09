package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoggerSetup(t *testing.T) {
	tests := []struct {
		name     string
		level    Level
		expected string
	}{
		{"debug level", LevelDebug, "debug"},
		{"info level", LevelInfo, "info"},
		{"warn level", LevelWarn, "warn"},
		{"error level", LevelError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cfg := &Config{
				Level:  tt.level,
				Format: "text",
				Output: &buf,
			}

			Setup(cfg)

			// Test that we can log at the configured level
			Info("test message")
			output := buf.String()

			if tt.level == LevelWarn || tt.level == LevelError {
				// Info messages should not appear at warn or error level
				if strings.Contains(output, "test message") {
					t.Errorf("Expected no output at %s level, got: %s", tt.level, output)
				}
			} else {
				// Info messages should appear at debug and info levels
				if !strings.Contains(output, "test message") {
					t.Errorf("Expected 'test message' in output, got: %s", output)
				}
			}
		})
	}
}

func TestParseLevelFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{"info", LevelInfo},
		{"INFO", LevelInfo},
		{"warn", LevelWarn},
		{"WARN", LevelWarn},
		{"warning", LevelWarn},
		{"error", LevelError},
		{"ERROR", LevelError},
		{"invalid", LevelInfo}, // should default to info
		{"", LevelInfo},        // should default to info
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseLevelFromString(tt.input)
			if result != tt.expected {
				t.Errorf("ParseLevelFromString(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

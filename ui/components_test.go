package ui

import (
	"strings"
	"testing"
)

func TestHeaderComponent(t *testing.T) {
	tests := []struct {
		name     string
		header   *HeaderComponent
		contains []string
	}{
		{
			name:     "simple header",
			header:   NewHeader("Test Header"),
			contains: []string{"Test Header"},
		},
		{
			name:     "header with emoji",
			header:   NewHeader("Test Header").WithEmoji("🎭"),
			contains: []string{"🎭", "Test Header"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.header.Render()
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected header to contain %q, got: %s", expected, result)
				}
			}
		})
	}
}

func TestLabelValueComponent(t *testing.T) {
	lv := NewLabelValue("Label:", "Value")
	result := lv.Render()

	if !strings.Contains(result, "Label:") {
		t.Errorf("Expected label-value to contain label, got: %s", result)
	}
	if !strings.Contains(result, "Value") {
		t.Errorf("Expected label-value to contain value, got: %s", result)
	}
}

func TestStatusComponent(t *testing.T) {
	tests := []struct {
		status   string
		text     string
		expected string
	}{
		{"success", "Operation completed", "Operation completed"},
		{"error", "Operation failed", "Operation failed"},
		{"warning", "Operation skipped", "Operation skipped"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			status := NewStatus(tt.status, tt.text)
			result := status.Render()

			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected status to contain %q, got: %s", tt.expected, result)
			}
		})
	}
}

func TestStatusComponentWithIcon(t *testing.T) {
	status := NewStatus("success", "Test").WithIcon("✓")
	result := status.Render()

	if !strings.Contains(result, "✓") {
		t.Errorf("Expected status to contain icon, got: %s", result)
	}
	if !strings.Contains(result, "Test") {
		t.Errorf("Expected status to contain text, got: %s", result)
	}
}

func TestListComponent(t *testing.T) {
	items := []string{"Item 1", "Item 2", "Item 3"}
	list := NewList(items)
	result := list.Render()

	for _, item := range items {
		if !strings.Contains(result, item) {
			t.Errorf("Expected list to contain %q, got: %s", item, result)
		}
	}

	// Should contain bullet points
	if !strings.Contains(result, "•") {
		t.Errorf("Expected list to contain bullet points, got: %s", result)
	}
}

func TestBoxComponent(t *testing.T) {
	box := NewBox("Test content").WithTitle("Test Title")
	result := box.Render()

	if !strings.Contains(result, "Test content") {
		t.Errorf("Expected box to contain content, got: %s", result)
	}
	if !strings.Contains(result, "Test Title") {
		t.Errorf("Expected box to contain title, got: %s", result)
	}
}

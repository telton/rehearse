package ui

import (
	"strings"
	"testing"
)

func TestTableRenderer(t *testing.T) {
	table := NewTable().
		AddColumn("Name", 15, "left").
		AddColumn("Status", 10, "center").
		AddColumn("Count", 8, "right").
		AddRow("job1", "success", "5").
		AddRow("job2", "failed", "0").
		AddRow("job3", "skipped", "3")

	result := table.Render()

	// Check headers are present
	if !strings.Contains(result, "Name") {
		t.Errorf("Expected table to contain Name header, got: %s", result)
	}
	if !strings.Contains(result, "Status") {
		t.Errorf("Expected table to contain Status header, got: %s", result)
	}
	if !strings.Contains(result, "Count") {
		t.Errorf("Expected table to contain Count header, got: %s", result)
	}

	// Check data rows are present
	if !strings.Contains(result, "job1") {
		t.Errorf("Expected table to contain job1 data, got: %s", result)
	}
	if !strings.Contains(result, "success") {
		t.Errorf("Expected table to contain success status, got: %s", result)
	}
	if !strings.Contains(result, "failed") {
		t.Errorf("Expected table to contain failed status, got: %s", result)
	}

	// Check separator is present
	if !strings.Contains(result, "─") {
		t.Errorf("Expected table to contain separator, got: %s", result)
	}
}

func TestTableRenderer_Empty(t *testing.T) {
	table := NewTable()
	result := table.Render()

	if result != "" {
		t.Errorf("Expected empty table to return empty string, got: %s", result)
	}
}

func TestProgressBar(t *testing.T) {
	progress := NewProgressBar(10).SetProgress(3)
	result := progress.Render()

	// Should contain some progress indication
	if !strings.Contains(result, "30%") {
		t.Errorf("Expected progress bar to contain percentage, got: %s", result)
	}
	if !strings.Contains(result, "(3/10)") {
		t.Errorf("Expected progress bar to contain count, got: %s", result)
	}
}

func TestProgressBar_ZeroTotal(t *testing.T) {
	progress := NewProgressBar(0)
	result := progress.Render()

	if result != "" {
		t.Errorf("Expected progress bar with zero total to return empty string, got: %s", result)
	}
}

func TestProgressBar_Complete(t *testing.T) {
	progress := NewProgressBar(5).SetProgress(5)
	result := progress.Render()

	if !strings.Contains(result, "100%") {
		t.Errorf("Expected complete progress bar to show 100%%, got: %s", result)
	}
}

func TestProgressBar_CustomWidth(t *testing.T) {
	progress := NewProgressBar(10).WithWidth(20).SetProgress(5)
	result := progress.Render()

	if !strings.Contains(result, "50%") {
		t.Errorf("Expected progress bar to show 50%%, got: %s", result)
	}
}

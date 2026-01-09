package ui

import (
	"strings"
	"testing"
)

func TestWorkflowRenderer(t *testing.T) {
	renderer := NewWorkflowRenderer()

	t.Run("RenderWorkflowHeader", func(t *testing.T) {
		result := renderer.RenderWorkflowHeader("test-workflow", "push")

		if !strings.Contains(result, "test-workflow") {
			t.Errorf("Expected header to contain workflow name, got: %s", result)
		}
		if !strings.Contains(result, "push") {
			t.Errorf("Expected header to contain trigger, got: %s", result)
		}
		if !strings.Contains(result, "🎭") {
			t.Errorf("Expected header to contain emoji, got: %s", result)
		}
	})

	t.Run("RenderContext", func(t *testing.T) {
		contextData := map[string]string{
			"ref":        "refs/heads/main",
			"event_name": "push",
			"sha":        "abc123",
		}
		result := renderer.RenderContext(contextData)

		if !strings.Contains(result, "Context:") {
			t.Errorf("Expected context to contain header, got: %s", result)
		}
		for key, value := range contextData {
			if !strings.Contains(result, key) {
				t.Errorf("Expected context to contain key %q, got: %s", key, result)
			}
			if !strings.Contains(result, value) {
				t.Errorf("Expected context to contain value %q, got: %s", value, result)
			}
		}
	})

	t.Run("RenderJobHeader", func(t *testing.T) {
		result := renderer.RenderJobHeader("build", "Build Application")

		if !strings.Contains(result, "build") {
			t.Errorf("Expected job header to contain job ID, got: %s", result)
		}
		if !strings.Contains(result, "Build Application") {
			t.Errorf("Expected job header to contain job name, got: %s", result)
		}
		if !strings.Contains(result, "🔧") {
			t.Errorf("Expected job header to contain emoji, got: %s", result)
		}
	})

	t.Run("RenderStep", func(t *testing.T) {
		tests := []struct {
			status       string
			expectedIcon string
		}{
			{"success", "✓"},
			{"error", "✗"},
			{"skipped", "⊝"},
			{"running", "⟳"},
		}

		for _, tt := range tests {
			result := renderer.RenderStep("Test Step", tt.status, 0)

			if !strings.Contains(result, "Test Step") {
				t.Errorf("Expected step to contain step name, got: %s", result)
			}
			if !strings.Contains(result, tt.expectedIcon) {
				t.Errorf("Expected step with status %q to contain icon %q, got: %s",
					tt.status, tt.expectedIcon, result)
			}
		}
	})

	t.Run("RenderSummary", func(t *testing.T) {
		result := renderer.RenderSummary(5, 3, 1, 1)

		if !strings.Contains(result, "Summary") {
			t.Errorf("Expected summary to contain header, got: %s", result)
		}
		if !strings.Contains(result, "📊") {
			t.Errorf("Expected summary to contain emoji, got: %s", result)
		}
		if !strings.Contains(result, "5") {
			t.Errorf("Expected summary to contain total count, got: %s", result)
		}
	})

	t.Run("RenderDockerOperation", func(t *testing.T) {
		result := renderer.RenderDockerOperation("pull", "ubuntu:latest")

		if !strings.Contains(result, "pull") {
			t.Errorf("Expected docker operation to contain operation, got: %s", result)
		}
		if !strings.Contains(result, "ubuntu:latest") {
			t.Errorf("Expected docker operation to contain image, got: %s", result)
		}
		if !strings.Contains(result, "🐳") {
			t.Errorf("Expected docker operation to contain emoji, got: %s", result)
		}
	})

	t.Run("RenderEnvironmentVar", func(t *testing.T) {
		result := renderer.RenderEnvironmentVar("TEST_VAR", "test_value")

		if !strings.Contains(result, "TEST_VAR") {
			t.Errorf("Expected env var to contain key, got: %s", result)
		}
		if !strings.Contains(result, "test_value") {
			t.Errorf("Expected env var to contain value, got: %s", result)
		}
		if !strings.Contains(result, "export") {
			t.Errorf("Expected env var to contain export keyword, got: %s", result)
		}
	})

	t.Run("RenderExpression", func(t *testing.T) {
		result := renderer.RenderExpression("${{ github.ref }}", "refs/heads/main")

		if !strings.Contains(result, "${{ github.ref }}") {
			t.Errorf("Expected expression to contain expression text, got: %s", result)
		}
		if !strings.Contains(result, "refs/heads/main") {
			t.Errorf("Expected expression to contain result, got: %s", result)
		}
		if !strings.Contains(result, "→") {
			t.Errorf("Expected expression to contain arrow, got: %s", result)
		}
	})
}

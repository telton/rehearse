package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShellStepExecutor_evaluateExpressions(t *testing.T) {
	executor := CreateTestShellExecutor(NewMockDockerClient())
	runtime := createTestRuntimeWithOutputs()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no expressions",
			input:    "echo 'Hello World'",
			expected: "echo 'Hello World'",
		},
		{
			name:     "single step output expression",
			input:    "echo 'Version: ${{ steps.version.outputs.number }}'",
			expected: "echo 'Version: 1.2.3'",
		},
		{
			name:     "multiple step output expressions",
			input:    "echo '${{ steps.version.outputs.number }} - ${{ steps.build.outputs.status }}'",
			expected: "echo '1.2.3 - success'",
		},
		{
			name:     "environment variable expression",
			input:    "echo 'Stage: ${{ env.DEPLOY_STAGE }}'",
			expected: "echo 'Stage: production'",
		},
		{
			name:     "mixed expressions",
			input:    "echo '${{ env.APP_NAME }} v${{ steps.version.outputs.number }} (${{ steps.build.outputs.status }})'",
			expected: "echo 'MyApp v1.2.3 (success)'",
		},
		{
			name:     "unresolved step output",
			input:    "echo '${{ steps.nonexistent.outputs.value }}'",
			expected: "echo ''",
		},
		{
			name:     "unresolved environment variable",
			input:    "echo '${{ env.NONEXISTENT_VAR }}'",
			expected: "echo ''",
		},
		{
			name:     "expression with whitespace",
			input:    "echo '${{  steps.version.outputs.number  }}'",
			expected: "echo '1.2.3'",
		},
		{
			name:     "nested braces (not expressions)",
			input:    "echo '{{ not an expression }}'",
			expected: "echo '{{ not an expression }}'",
		},
		{
			name:     "malformed expression",
			input:    "echo '${{ steps.version.outputs }}'",
			expected: "echo ''",
		},
		{
			name:     "complex multi-line with expressions",
			input:    "echo 'App: ${{ env.APP_NAME }}'\necho 'Version: ${{ steps.version.outputs.number }}'",
			expected: "echo 'App: MyApp'\necho 'Version: 1.2.3'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.evaluateExpressions(tt.input, runtime)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShellStepExecutor_evaluateExpression(t *testing.T) {
	executor := CreateTestShellExecutor(NewMockDockerClient())
	runtime := createTestRuntimeWithOutputs()

	tests := []struct {
		name       string
		expression string
		expected   string
	}{
		{
			name:       "valid step output",
			expression: "steps.version.outputs.number",
			expected:   "1.2.3",
		},
		{
			name:       "valid step output with different step",
			expression: "steps.build.outputs.status",
			expected:   "success",
		},
		{
			name:       "valid environment variable",
			expression: "env.APP_NAME",
			expected:   "MyApp",
		},
		{
			name:       "valid environment variable different var",
			expression: "env.DEPLOY_STAGE",
			expected:   "production",
		},
		{
			name:       "nonexistent step",
			expression: "steps.missing.outputs.value",
			expected:   "",
		},
		{
			name:       "nonexistent output",
			expression: "steps.version.outputs.missing",
			expected:   "",
		},
		{
			name:       "nonexistent environment variable",
			expression: "env.MISSING_VAR",
			expected:   "",
		},
		{
			name:       "malformed step expression - missing outputs",
			expression: "steps.version.number",
			expected:   "",
		},
		{
			name:       "malformed step expression - too few parts",
			expression: "steps.version",
			expected:   "",
		},
		{
			name:       "malformed env expression",
			expression: "environment.VAR",
			expected:   "",
		},
		{
			name:       "unknown expression type",
			expression: "unknown.expression.type",
			expected:   "",
		},
		{
			name:       "empty expression",
			expression: "",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.evaluateExpression(tt.expression, runtime)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExecutor_evaluateOutputExpression(t *testing.T) {
	mockDocker := NewMockDockerClient()
	mockGit := NewMockGitRepo()
	executor := NewExecutor(&Analyzer{}, mockDocker, mockGit)

	executor.runtime = createTestRuntimeWithOutputs()

	tests := []struct {
		name       string
		expression string
		expected   string
	}{
		{
			name:       "step output with wrapper",
			expression: "${{ steps.version.outputs.number }}",
			expected:   "1.2.3",
		},
		{
			name:       "step output without wrapper",
			expression: "steps.version.outputs.number",
			expected:   "1.2.3",
		},
		{
			name:       "environment variable with wrapper",
			expression: "${{ env.APP_NAME }}",
			expected:   "MyApp",
		},
		{
			name:       "environment variable without wrapper",
			expression: "env.DEPLOY_STAGE",
			expected:   "production",
		},
		{
			name:       "nonexistent step output",
			expression: "${{ steps.missing.outputs.value }}",
			expected:   "",
		},
		{
			name:       "literal string (not an expression)",
			expression: "literal-value",
			expected:   "literal-value",
		},
		{
			name:       "complex expression with whitespace",
			expression: "${{  steps.build.outputs.status  }}",
			expected:   "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.evaluateOutputExpression(tt.expression)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// createTestRuntimeWithOutputs creates a runtime with test step outputs and environment variables.
func createTestRuntimeWithOutputs() *Runtime {
	runtime := &Runtime{
		DynamicEnv:  make(map[string]string),
		StepOutputs: make(map[string]map[string]string),
	}

	runtime.DynamicEnv["APP_NAME"] = "MyApp"
	runtime.DynamicEnv["DEPLOY_STAGE"] = "production"
	runtime.DynamicEnv["BUILD_NUMBER"] = "1234"

	runtime.StepOutputs["version"] = map[string]string{
		"number":     "1.2.3",
		"commit_sha": "abc123def",
	}

	runtime.StepOutputs["build"] = map[string]string{
		"status":   "success",
		"duration": "120s",
	}

	runtime.StepOutputs["test"] = map[string]string{
		"coverage":   "98.5",
		"test_count": "156",
	}

	return runtime
}

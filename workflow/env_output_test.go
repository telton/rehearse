package workflow

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutor_processStepOutputFiles_GITHUB_ENV(t *testing.T) {
	executor := NewExecutor(&Analyzer{}, NewMockDockerClient(), NewMockGitRepo())

	tempDir := t.TempDir()
	executor.runtime.TempDir = tempDir

	tests := []struct {
		name        string
		fileContent string
		expected    map[string]string
	}{
		{
			name:        "single environment variable",
			fileContent: "NODE_VERSION=18.20.0",
			expected:    map[string]string{"NODE_VERSION": "18.20.0"},
		},
		{
			name:        "multiple environment variables",
			fileContent: "NODE_VERSION=18.20.0\nBUILD_NUMBER=1234\nDEPLOY_ENV=staging",
			expected: map[string]string{
				"NODE_VERSION": "18.20.0",
				"BUILD_NUMBER": "1234",
				"DEPLOY_ENV":   "staging",
			},
		},
		{
			name:        "empty lines and whitespace",
			fileContent: "\nNODE_VERSION=18.20.0\n\n  BUILD_NUMBER=1234  \n\n",
			expected: map[string]string{
				"NODE_VERSION": "18.20.0",
				"BUILD_NUMBER": "1234",
			},
		},
		{
			name:        "values with spaces and special characters",
			fileContent: "APP_NAME=My Application\nVERSION=v1.2.3-beta\nPATH=/usr/bin:/bin",
			expected: map[string]string{
				"APP_NAME": "My Application",
				"VERSION":  "v1.2.3-beta",
				"PATH":     "/usr/bin:/bin",
			},
		},
		{
			name:        "empty file",
			fileContent: "",
			expected:    map[string]string{},
		},
		{
			name:        "malformed lines ignored",
			fileContent: "VALID_VAR=value\nINVALID_LINE\nANOTHER_VALID=test",
			expected: map[string]string{
				"VALID_VAR":     "value",
				"ANOTHER_VALID": "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor.runtime.DynamicEnv = make(map[string]string)

			envFile := filepath.Join(tempDir, "GITHUB_ENV")
			err := os.WriteFile(envFile, []byte(tt.fileContent), 0600)
			require.NoError(t, err)

			err = executor.processStepOutputFiles("test-step")
			assert.NoError(t, err)

			assert.Equal(t, tt.expected, executor.runtime.DynamicEnv)

			content, err := os.ReadFile(envFile)
			assert.NoError(t, err)
			assert.Empty(t, content)
		})
	}
}

func TestExecutor_processStepOutputFiles_GITHUB_OUTPUT(t *testing.T) {
	executor := NewExecutor(&Analyzer{}, NewMockDockerClient(), NewMockGitRepo())

	tempDir := t.TempDir()

	executor.runtime.TempDir = tempDir

	tests := []struct {
		name        string
		stepID      string
		fileContent string
		expected    map[string]string
	}{
		{
			name:        "single output",
			stepID:      "build",
			fileContent: "version=1.2.3",
			expected:    map[string]string{"version": "1.2.3"},
		},
		{
			name:        "multiple outputs",
			stepID:      "test",
			fileContent: "status=passed\ncoverage=98.5\nduration=120s",
			expected: map[string]string{
				"status":   "passed",
				"coverage": "98.5",
				"duration": "120s",
			},
		},
		{
			name:        "outputs with complex values",
			stepID:      "deploy",
			fileContent: "url=https://app.example.com\ncommit_sha=abc123def456\ntimestamp=2026-01-09T16:30:00Z",
			expected: map[string]string{
				"url":        "https://app.example.com",
				"commit_sha": "abc123def456",
				"timestamp":  "2026-01-09T16:30:00Z",
			},
		},
		{
			name:        "empty file",
			stepID:      "empty",
			fileContent: "",
			expected:    map[string]string{},
		},
		{
			name:        "whitespace handling",
			stepID:      "whitespace",
			fileContent: "\n  version=1.2.3  \n\n  status=success  \n",
			expected: map[string]string{
				"version": "1.2.3",
				"status":  "success",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor.runtime.StepOutputs = make(map[string]map[string]string)

			outputFile := filepath.Join(tempDir, "GITHUB_OUTPUT")
			err := os.WriteFile(outputFile, []byte(tt.fileContent), 0600)
			require.NoError(t, err)

			err = executor.processStepOutputFiles(tt.stepID)
			assert.NoError(t, err)

			if len(tt.expected) > 0 {
				assert.Contains(t, executor.runtime.StepOutputs, tt.stepID)
				assert.Equal(t, tt.expected, executor.runtime.StepOutputs[tt.stepID])
			} else {
				if outputs, exists := executor.runtime.StepOutputs[tt.stepID]; exists {
					assert.Empty(t, outputs)
				}
			}

			content, err := os.ReadFile(outputFile)
			assert.NoError(t, err)
			assert.Empty(t, content)
		})
	}
}

func TestExecutor_processStepOutputFiles_NullByteHandling(t *testing.T) {
	executor := NewExecutor(&Analyzer{}, NewMockDockerClient(), NewMockGitRepo())

	tempDir := t.TempDir()

	executor.runtime.TempDir = tempDir

	envContent := "VAR1=value1\x00\x00\nVAR2=value2\x00"
	envFile := filepath.Join(tempDir, "GITHUB_ENV")
	err := os.WriteFile(envFile, []byte(envContent), 0600)
	require.NoError(t, err)

	outputContent := "output1=test\x00\x00\noutput2=value\x00"
	outputFile := filepath.Join(tempDir, "GITHUB_OUTPUT")
	err = os.WriteFile(outputFile, []byte(outputContent), 0600)
	require.NoError(t, err)

	err = executor.processStepOutputFiles("test")
	assert.NoError(t, err)

	expectedEnv := map[string]string{
		"VAR1": "value1",
		"VAR2": "value2",
	}
	assert.Equal(t, expectedEnv, executor.runtime.DynamicEnv)

	expectedOutput := map[string]string{
		"output1": "test",
		"output2": "value",
	}
	assert.Equal(t, expectedOutput, executor.runtime.StepOutputs["test"])
}

func TestExecutor_processStepOutputFiles_NoTempDir(t *testing.T) {
	executor := NewExecutor(&Analyzer{}, NewMockDockerClient(), NewMockGitRepo())
	executor.runtime.TempDir = ""

	err := executor.processStepOutputFiles("test")
	assert.NoError(t, err)

	assert.Empty(t, executor.runtime.DynamicEnv)
	assert.Empty(t, executor.runtime.StepOutputs)
}

func TestExecutor_processStepOutputFiles_MissingFiles(t *testing.T) {
	executor := NewExecutor(&Analyzer{}, NewMockDockerClient(), NewMockGitRepo())

	tempDir := t.TempDir()

	executor.runtime.TempDir = tempDir

	err := executor.processStepOutputFiles("test")
	assert.NoError(t, err)

	assert.Empty(t, executor.runtime.DynamicEnv)
	assert.Empty(t, executor.runtime.StepOutputs)
}

package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecutor_processJobOutputs(t *testing.T) {
	mockDocker := NewMockDockerClient()
	mockGit := NewMockGitRepo()
	executor := NewExecutor(&Analyzer{}, mockDocker, mockGit)

	tests := []struct {
		name            string
		job             *Job
		stepOutputs     map[string]map[string]string
		dynamicEnv      map[string]string
		expectedOutputs map[string]string
	}{
		{
			name: "simple step output reference",
			job: &Job{
				Name: "test-job",
				Outputs: map[string]string{
					"version": "${{ steps.build.outputs.version }}",
				},
			},
			stepOutputs: map[string]map[string]string{
				"build": {"version": "1.2.3"},
			},
			dynamicEnv: map[string]string{},
			expectedOutputs: map[string]string{
				"version": "1.2.3",
			},
		},
		{
			name: "multiple step output references",
			job: &Job{
				Name: "complex-job",
				Outputs: map[string]string{
					"app-version":   "${{ steps.version.outputs.number }}",
					"build-status":  "${{ steps.build.outputs.status }}",
					"test-coverage": "${{ steps.test.outputs.coverage }}",
				},
			},
			stepOutputs: map[string]map[string]string{
				"version": {"number": "2.1.0"},
				"build":   {"status": "success", "duration": "120s"},
				"test":    {"coverage": "98.5", "passed": "156"},
			},
			dynamicEnv: map[string]string{},
			expectedOutputs: map[string]string{
				"app-version":   "2.1.0",
				"build-status":  "success",
				"test-coverage": "98.5",
			},
		},
		{
			name: "environment variable reference",
			job: &Job{
				Name: "env-job",
				Outputs: map[string]string{
					"deploy-env": "${{ env.DEPLOY_ENVIRONMENT }}",
					"app-name":   "${{ env.APPLICATION_NAME }}",
				},
			},
			stepOutputs: map[string]map[string]string{},
			dynamicEnv: map[string]string{
				"DEPLOY_ENVIRONMENT": "production",
				"APPLICATION_NAME":   "MyApp",
			},
			expectedOutputs: map[string]string{
				"deploy-env": "production",
				"app-name":   "MyApp",
			},
		},
		{
			name: "mixed step outputs and environment variables",
			job: &Job{
				Name: "mixed-job",
				Outputs: map[string]string{
					"version":    "${{ steps.build.outputs.version }}",
					"deploy-env": "${{ env.DEPLOY_ENV }}",
					"status":     "${{ steps.test.outputs.result }}",
				},
			},
			stepOutputs: map[string]map[string]string{
				"build": {"version": "3.0.0", "timestamp": "2026-01-09"},
				"test":  {"result": "passed", "duration": "45s"},
			},
			dynamicEnv: map[string]string{
				"DEPLOY_ENV": "staging",
				"BUILD_ENV":  "ci",
			},
			expectedOutputs: map[string]string{
				"version":    "3.0.0",
				"deploy-env": "staging",
				"status":     "passed",
			},
		},
		{
			name: "nonexistent step output",
			job: &Job{
				Name: "missing-job",
				Outputs: map[string]string{
					"missing-output": "${{ steps.nonexistent.outputs.value }}",
					"valid-output":   "${{ steps.build.outputs.version }}",
				},
			},
			stepOutputs: map[string]map[string]string{
				"build": {"version": "1.0.0"},
			},
			dynamicEnv: map[string]string{},
			expectedOutputs: map[string]string{
				"missing-output": "", // Should be empty for unresolved expression
				"valid-output":   "1.0.0",
			},
		},
		{
			name: "nonexistent environment variable",
			job: &Job{
				Name: "missing-env-job",
				Outputs: map[string]string{
					"missing-env": "${{ env.NONEXISTENT_VAR }}",
					"valid-env":   "${{ env.EXISTING_VAR }}",
				},
			},
			stepOutputs: map[string]map[string]string{},
			dynamicEnv: map[string]string{
				"EXISTING_VAR": "existing_value",
			},
			expectedOutputs: map[string]string{
				"missing-env": "", // Should be empty for unresolved expression
				"valid-env":   "existing_value",
			},
		},
		{
			name: "literal values (not expressions)",
			job: &Job{
				Name: "literal-job",
				Outputs: map[string]string{
					"literal": "static-value",
					"version": "${{ steps.build.outputs.version }}",
					"another": "another-static-value",
				},
			},
			stepOutputs: map[string]map[string]string{
				"build": {"version": "4.0.0"},
			},
			dynamicEnv: map[string]string{},
			expectedOutputs: map[string]string{
				"literal": "static-value",
				"version": "4.0.0",
				"another": "another-static-value",
			},
		},
		{
			name: "no outputs defined",
			job: &Job{
				Name:    "no-outputs-job",
				Outputs: nil,
			},
			stepOutputs:     map[string]map[string]string{},
			dynamicEnv:      map[string]string{},
			expectedOutputs: map[string]string{}, // Should remain empty
		},
		{
			name: "empty outputs map",
			job: &Job{
				Name:    "empty-outputs-job",
				Outputs: map[string]string{},
			},
			stepOutputs:     map[string]map[string]string{},
			dynamicEnv:      map[string]string{},
			expectedOutputs: map[string]string{}, // Should remain empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor.runtime.JobContext = &ExecutionJobContext{
				Job:     tt.job,
				Outputs: make(map[string]string),
			}
			executor.runtime.StepOutputs = tt.stepOutputs
			executor.runtime.DynamicEnv = tt.dynamicEnv

			executor.processJobOutputs(tt.job)

			assert.Equal(t, tt.expectedOutputs, executor.runtime.JobContext.Outputs,
				"Job outputs don't match expected values")
		})
	}
}

func TestExecutor_processJobOutputs_WithExpressionWhitespace(t *testing.T) {
	mockDocker := NewMockDockerClient()
	mockGit := NewMockGitRepo()
	executor := NewExecutor(&Analyzer{}, mockDocker, mockGit)

	job := &Job{
		Name: "whitespace-job",
		Outputs: map[string]string{
			"normal":       "${{ steps.build.outputs.version }}",
			"extra-spaces": "${{  steps.build.outputs.version  }}",
			"many-spaces":  "${{    steps.build.outputs.version    }}",
		},
	}

	stepOutputs := map[string]map[string]string{
		"build": {"version": "1.2.3"},
	}

	executor.runtime.JobContext = &ExecutionJobContext{
		Job:     job,
		Outputs: make(map[string]string),
	}
	executor.runtime.StepOutputs = stepOutputs
	executor.runtime.DynamicEnv = make(map[string]string)

	executor.processJobOutputs(job)

	expectedValue := "1.2.3"
	assert.Equal(t, expectedValue, executor.runtime.JobContext.Outputs["normal"])
	assert.Equal(t, expectedValue, executor.runtime.JobContext.Outputs["extra-spaces"])
	assert.Equal(t, expectedValue, executor.runtime.JobContext.Outputs["many-spaces"])
}

func TestExecutor_processJobOutputs_EmptyJobContext(t *testing.T) {
	mockDocker := NewMockDockerClient()
	mockGit := NewMockGitRepo()
	executor := NewExecutor(&Analyzer{}, mockDocker, mockGit)

	executor.runtime.JobContext = nil

	job := &Job{
		Name: "test-job",
		Outputs: map[string]string{
			"version": "${{ steps.build.outputs.version }}",
		},
	}

	assert.NotPanics(t, func() {
		executor.processJobOutputs(job)
	})
}

func TestExecutor_processJobOutputs_Integration(t *testing.T) {
	mockDocker := NewMockDockerClient()
	mockGit := NewMockGitRepo()
	executor := NewExecutor(&Analyzer{}, mockDocker, mockGit)

	job := &Job{
		Name: "integration-job",
		Outputs: map[string]string{
			"app-version":   "${{ steps.version.outputs.number }}",
			"build-result":  "${{ steps.build.outputs.status }}",
			"deploy-target": "${{ env.DEPLOY_TARGET }}",
			"test-summary":  "${{ steps.test.outputs.summary }}",
		},
	}

	executor.runtime.JobContext = &ExecutionJobContext{
		Job:     job,
		Outputs: make(map[string]string),
	}

	executor.runtime.StepOutputs = map[string]map[string]string{
		"version": {
			"number":     "2.1.0",
			"commit_sha": "abc123",
		},
		"build": {
			"status":   "success",
			"duration": "180s",
			"size":     "45MB",
		},
		"test": {
			"summary":  "156 passed, 0 failed",
			"coverage": "97.2%",
		},
	}

	executor.runtime.DynamicEnv = map[string]string{
		"DEPLOY_TARGET": "production",
		"BUILD_ENV":     "ci",
	}

	executor.processJobOutputs(job)

	expectedOutputs := map[string]string{
		"app-version":   "2.1.0",
		"build-result":  "success",
		"deploy-target": "production",
		"test-summary":  "156 passed, 0 failed",
	}

	assert.Equal(t, expectedOutputs, executor.runtime.JobContext.Outputs)
}

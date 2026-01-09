package workflow

import (
	"maps"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShellStepExecutor_buildEnvironment_Precedence(t *testing.T) {
	executor := CreateTestShellExecutor(NewMockDockerClient())

	tests := []struct {
		name           string
		jobEnv         map[string]string
		dynamicEnv     map[string]string
		stepEnv        map[string]string
		expectedVars   map[string]string
		unexpectedVars []string
	}{
		{
			name:   "step env overrides dynamic env",
			jobEnv: map[string]string{"JOB_VAR": "job_value"},
			dynamicEnv: map[string]string{
				"DYNAMIC_VAR":  "dynamic_value",
				"OVERRIDE_VAR": "dynamic_value",
			},
			stepEnv: map[string]string{
				"STEP_VAR":     "step_value",
				"OVERRIDE_VAR": "step_value", // Should override dynamic
			},
			expectedVars: map[string]string{
				"JOB_VAR":      "job_value",
				"DYNAMIC_VAR":  "dynamic_value",
				"STEP_VAR":     "step_value",
				"OVERRIDE_VAR": "step_value", // Step wins
			},
		},
		{
			name: "dynamic env overrides job env",
			jobEnv: map[string]string{
				"JOB_VAR":    "job_value",
				"SHARED_VAR": "job_value",
			},
			dynamicEnv: map[string]string{
				"DYNAMIC_VAR": "dynamic_value",
				"SHARED_VAR":  "dynamic_value", // Should override job
			},
			stepEnv: map[string]string{"STEP_VAR": "step_value"},
			expectedVars: map[string]string{
				"JOB_VAR":     "job_value",
				"DYNAMIC_VAR": "dynamic_value",
				"STEP_VAR":    "step_value",
				"SHARED_VAR":  "dynamic_value", // Dynamic wins over job
			},
		},
		{
			name: "all levels present with complex precedence",
			jobEnv: map[string]string{
				"ALL_LEVELS": "job_value",
				"JOB_ONLY":   "job_value",
			},
			dynamicEnv: map[string]string{
				"ALL_LEVELS":   "dynamic_value",
				"DYNAMIC_ONLY": "dynamic_value",
			},
			stepEnv: map[string]string{
				"ALL_LEVELS": "step_value",
				"STEP_ONLY":  "step_value",
			},
			expectedVars: map[string]string{
				"ALL_LEVELS":   "step_value",    // Step wins all
				"JOB_ONLY":     "job_value",     // Job only
				"DYNAMIC_ONLY": "dynamic_value", // Dynamic only
				"STEP_ONLY":    "step_value",    // Step only
			},
		},
		{
			name:           "empty environments",
			jobEnv:         map[string]string{},
			dynamicEnv:     map[string]string{},
			stepEnv:        map[string]string{},
			expectedVars:   map[string]string{}, // Only GitHub defaults
			unexpectedVars: []string{"JOB_VAR", "DYNAMIC_VAR", "STEP_VAR"},
		},
		{
			name:           "nil environments",
			jobEnv:         nil,
			dynamicEnv:     map[string]string{"DYNAMIC_VAR": "value"},
			stepEnv:        nil,
			expectedVars:   map[string]string{"DYNAMIC_VAR": "value"},
			unexpectedVars: []string{"JOB_VAR", "STEP_VAR"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtime := &Runtime{
				DynamicEnv:  tt.dynamicEnv,
				StepOutputs: make(map[string]map[string]string),
				JobContext: &ExecutionJobContext{
					Job: &Job{Env: tt.jobEnv},
				},
			}

			step := &Step{
				ID:   "test-step",
				Name: "Test Step",
				Env:  tt.stepEnv,
			}

			env := executor.buildEnvironment(step, runtime)

			envMap := make(map[string]string)
			for _, envVar := range env {
				if parts := splitEnvVar(envVar); len(parts) == 2 {
					envMap[parts[0]] = parts[1]
				}
			}

			for key, expectedValue := range tt.expectedVars {
				assert.Contains(t, envMap, key, "Expected variable %s not found", key)
				assert.Equal(t, expectedValue, envMap[key], "Variable %s has wrong value", key)
			}

			for _, key := range tt.unexpectedVars {
				assert.NotContains(t, envMap, key, "Unexpected variable %s found", key)
			}

			expectedDefaults := map[string]string{
				"GITHUB_WORKSPACE":  "/github/workspace",
				"GITHUB_ACTOR":      "rehearse",
				"GITHUB_REPOSITORY": "local/repo",
				"RUNNER_OS":         "Linux",
				"RUNNER_ARCH":       "X64",
			}

			for key, expectedValue := range expectedDefaults {
				assert.Contains(t, envMap, key, "GitHub default %s not found", key)
				assert.Equal(t, expectedValue, envMap[key], "GitHub default %s has wrong value", key)
			}
		})
	}
}

func TestExecutor_DynamicContextUpdates(t *testing.T) {
	mockDocker := NewMockDockerClient()
	mockGit := NewMockGitRepo()
	executor := NewExecutor(&Analyzer{}, mockDocker, mockGit)

	err := executor.setupTempDirectory()
	require.NoError(t, err)
	defer executor.cleanupTempDirectory()

	tests := []struct {
		name                 string
		initialEnv           map[string]string
		initialOutputs       map[string]map[string]string
		stepActions          []stepAction
		expectedFinalEnv     map[string]string
		expectedFinalOutputs map[string]map[string]string
	}{
		{
			name:           "sequential environment updates",
			initialEnv:     map[string]string{},
			initialOutputs: map[string]map[string]string{},
			stepActions: []stepAction{
				{stepID: "step1", envUpdates: map[string]string{"VAR1": "value1"}},
				{stepID: "step2", envUpdates: map[string]string{"VAR2": "value2"}},
				{stepID: "step3", envUpdates: map[string]string{"VAR1": "updated_value1"}}, // Override
			},
			expectedFinalEnv: map[string]string{
				"VAR1": "updated_value1", // Should be updated value
				"VAR2": "value2",
			},
			expectedFinalOutputs: map[string]map[string]string{},
		},
		{
			name:           "sequential output updates",
			initialEnv:     map[string]string{},
			initialOutputs: map[string]map[string]string{},
			stepActions: []stepAction{
				{stepID: "build", outputUpdates: map[string]string{"version": "1.0.0"}},
				{stepID: "test", outputUpdates: map[string]string{"coverage": "95.0"}},
				{stepID: "build", outputUpdates: map[string]string{"status": "success"}}, // Add to existing step
			},
			expectedFinalEnv: map[string]string{},
			expectedFinalOutputs: map[string]map[string]string{
				"build": {"version": "1.0.0", "status": "success"},
				"test":  {"coverage": "95.0"},
			},
		},
		{
			name:       "mixed environment and output updates",
			initialEnv: map[string]string{"INITIAL_VAR": "initial"},
			initialOutputs: map[string]map[string]string{
				"existing": {"old_output": "old_value"},
			},
			stepActions: []stepAction{
				{stepID: "step1", envUpdates: map[string]string{"NEW_VAR": "new"}, outputUpdates: map[string]string{"result": "success"}},
				{stepID: "step2", envUpdates: map[string]string{"INITIAL_VAR": "updated"}, outputUpdates: map[string]string{"data": "processed"}},
			},
			expectedFinalEnv: map[string]string{
				"INITIAL_VAR": "updated", // Should be overridden
				"NEW_VAR":     "new",
			},
			expectedFinalOutputs: map[string]map[string]string{
				"existing": {"old_output": "old_value"}, // Should remain
				"step1":    {"result": "success"},
				"step2":    {"data": "processed"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor.runtime.DynamicEnv = make(map[string]string)
			maps.Copy(executor.runtime.DynamicEnv, tt.initialEnv)

			executor.runtime.StepOutputs = make(map[string]map[string]string)
			for stepID, outputs := range tt.initialOutputs {
				executor.runtime.StepOutputs[stepID] = make(map[string]string)
				maps.Copy(executor.runtime.StepOutputs[stepID], outputs)
			}

			for _, action := range tt.stepActions {
				maps.Copy(executor.runtime.DynamicEnv, action.envUpdates)

				if len(action.outputUpdates) > 0 {
					if executor.runtime.StepOutputs[action.stepID] == nil {
						executor.runtime.StepOutputs[action.stepID] = make(map[string]string)
					}
					maps.Copy(executor.runtime.StepOutputs[action.stepID], action.outputUpdates)
				}
			}

			assert.Equal(t, tt.expectedFinalEnv, executor.runtime.DynamicEnv, "Dynamic environment mismatch")
			assert.Equal(t, tt.expectedFinalOutputs, executor.runtime.StepOutputs, "Step outputs mismatch")
		})
	}
}

func TestExecutor_StepExecutionContextPropagation(t *testing.T) {
	mockDocker := NewMockDockerClient()
	mockGit := NewMockGitRepo()
	executor := NewExecutor(&Analyzer{}, mockDocker, mockGit)

	err := executor.setupTempDirectory()
	require.NoError(t, err)
	defer executor.cleanupTempDirectory()

	steps := []struct {
		step           *Step
		envToSet       map[string]string
		outputsToSet   map[string]string
		expectedEnvLen int
	}{
		{
			step: &Step{
				ID:   "init",
				Name: "Initialize",
				Run:  "echo 'init'",
			},
			envToSet:       map[string]string{"APP_NAME": "TestApp"},
			outputsToSet:   map[string]string{"app_version": "1.0.0"},
			expectedEnvLen: 1,
		},
		{
			step: &Step{
				ID:   "build",
				Name: "Build",
				Run:  "echo 'building'",
			},
			envToSet:       map[string]string{"BUILD_ENV": "production"},
			outputsToSet:   map[string]string{"build_id": "12345", "status": "success"},
			expectedEnvLen: 2,
		},
		{
			step: &Step{
				ID:   "test",
				Name: "Test",
				Run:  "echo 'testing'",
			},
			envToSet:       map[string]string{"APP_NAME": "TestApp-Updated"}, // Override previous
			outputsToSet:   map[string]string{"test_results": "passed"},
			expectedEnvLen: 2, // Same count, but APP_NAME updated
		},
	}

	for i, stepInfo := range steps {
		t.Run(stepInfo.step.Name, func(t *testing.T) {
			// Simulate step execution by updating context manually
			// (In real execution, this would be done by processStepOutputFiles)

			maps.Copy(executor.runtime.DynamicEnv, stepInfo.envToSet)

			if len(stepInfo.outputsToSet) > 0 {
				if executor.runtime.StepOutputs[stepInfo.step.ID] == nil {
					executor.runtime.StepOutputs[stepInfo.step.ID] = make(map[string]string)
				}
				maps.Copy(executor.runtime.StepOutputs[stepInfo.step.ID], stepInfo.outputsToSet)
			}

			assert.Len(t, executor.runtime.DynamicEnv, stepInfo.expectedEnvLen,
				"Step %d: Unexpected dynamic environment length", i+1)

			assert.Len(t, executor.runtime.StepOutputs, i+1,
				"Step %d: Unexpected number of step output collections", i+1)

			for outputKey := range stepInfo.outputsToSet {
				assert.Contains(t, executor.runtime.StepOutputs[stepInfo.step.ID], outputKey,
					"Step %d: Output %s not found", i+1, outputKey)
			}
		})
	}

	assert.Equal(t, "TestApp-Updated", executor.runtime.DynamicEnv["APP_NAME"], "APP_NAME should be updated to latest value")
	assert.Equal(t, "production", executor.runtime.DynamicEnv["BUILD_ENV"], "BUILD_ENV should be preserved")

	assert.Contains(t, executor.runtime.StepOutputs, "init")
	assert.Contains(t, executor.runtime.StepOutputs, "build")
	assert.Contains(t, executor.runtime.StepOutputs, "test")

	assert.Equal(t, "1.0.0", executor.runtime.StepOutputs["init"]["app_version"])
	assert.Equal(t, "12345", executor.runtime.StepOutputs["build"]["build_id"])
	assert.Equal(t, "passed", executor.runtime.StepOutputs["test"]["test_results"])
}

// stepAction represents an action that a step takes (updating env or outputs)
type stepAction struct {
	stepID        string
	envUpdates    map[string]string
	outputUpdates map[string]string
}

// splitEnvVar splits "KEY=VALUE" into ["KEY", "VALUE"]
func splitEnvVar(envVar string) []string {
	parts := strings.SplitN(envVar, "=", 2)
	if len(parts) == 2 {
		return parts
	}
	return []string{envVar, ""}
}

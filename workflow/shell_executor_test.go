package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShellStepExecutor_CanExecute(t *testing.T) {
	executor := &ShellStepExecutor{}

	tests := []struct {
		name     string
		step     *Step
		expected bool
	}{
		{
			name: "step with run command",
			step: &Step{
				ID:   "test",
				Name: "Test Step",
				Run:  "echo 'hello'",
			},
			expected: true,
		},
		{
			name: "step with uses action",
			step: &Step{
				ID:   "test",
				Name: "Test Step",
				Uses: "actions/checkout@v4",
			},
			expected: false,
		},
		{
			name: "step with empty run",
			step: &Step{
				ID:   "test",
				Name: "Test Step",
				Run:  "",
			},
			expected: false,
		},
		{
			name: "step with both run and uses",
			step: &Step{
				ID:   "test",
				Name: "Test Step",
				Run:  "echo 'hello'",
				Uses: "actions/checkout@v4",
			},
			expected: true, // run takes precedence
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.CanExecute(tt.step)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShellStepExecutor_Execute_BasicCommand(t *testing.T) {
	mockDocker := NewMockDockerClient()
	executor := CreateTestShellExecutor(mockDocker)

	step := CreateTestStep("echo-step", "Echo Step", "echo 'Hello World'")
	runtime := CreateTestRuntime("/tmp/workspace")

	expectedConfig := &ContainerConfig{
		Image:      "ubuntu:latest",
		Cmd:        []string{"sh", "-c", "echo 'Hello World'"},
		WorkingDir: "/github/workspace",
		Volumes: []VolumeMount{
			{Source: "/tmp/workspace", Target: "/github/workspace", Type: "bind"},
		},
	}

	mockDocker.On("PullImage", mock.Anything, "ubuntu:latest").Return(nil)
	mockDocker.On("CreateContainer", mock.Anything, mock.MatchedBy(func(config *ContainerConfig) bool {
		return config.Image == expectedConfig.Image &&
			len(config.Cmd) == 3 &&
			config.Cmd[0] == "sh" &&
			config.Cmd[1] == "-c" &&
			config.Cmd[2] == "echo 'Hello World'" &&
			config.WorkingDir == expectedConfig.WorkingDir &&
			len(config.Volumes) == 1 &&
			config.Volumes[0].Source == "/tmp/workspace" &&
			config.Volumes[0].Target == "/github/workspace"
	})).Return("container-123", nil)

	mockDocker.On("StartContainer", mock.Anything, "container-123").Return(nil)
	mockDocker.On("StopContainer", mock.Anything, "container-123").Return(nil)
	mockDocker.On("RemoveContainer", mock.Anything, "container-123").Return(nil)

	ctx := t.Context()
	result, err := executor.Execute(ctx, step, runtime)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, 0, result.ExitCode)
	assert.NotNil(t, result.Outputs)

	mockDocker.AssertExpectations(t)

	assert.Empty(t, runtime.Containers)
}

func TestShellStepExecutor_Execute_CustomContainer(t *testing.T) {
	mockDocker := NewMockDockerClient()
	executor := CreateTestShellExecutor(mockDocker)

	step := CreateTestStep("node-step", "Node Step", "npm test")
	runtime := CreateTestRuntime("/tmp/workspace")

	runtime.JobContext.Job.Container = &Container{
		Image: "node:18",
		Env:   map[string]string{"NODE_ENV": "test"},
	}

	mockDocker.On("PullImage", mock.Anything, "node:18").Return(nil)
	mockDocker.On("CreateContainer", mock.Anything, mock.MatchedBy(func(config *ContainerConfig) bool {
		return config.Image == "node:18" &&
			config.Cmd[2] == "npm test"
	})).Return("node-container", nil)

	mockDocker.On("StartContainer", mock.Anything, "node-container").Return(nil)
	mockDocker.On("StopContainer", mock.Anything, "node-container").Return(nil)
	mockDocker.On("RemoveContainer", mock.Anything, "node-container").Return(nil)

	ctx := t.Context()
	result, err := executor.Execute(ctx, step, runtime)

	assert.NoError(t, err)
	assert.True(t, result.Success)
	mockDocker.AssertExpectations(t)
}

func TestShellStepExecutor_Execute_ContainerCreationFailure(t *testing.T) {
	mockDocker := NewMockDockerClient()
	executor := CreateTestShellExecutor(mockDocker)

	step := CreateTestStep("failing-step", "Failing Step", "echo 'fail'")
	runtime := CreateTestRuntime("/tmp/workspace")

	mockDocker.On("PullImage", mock.Anything, "ubuntu:latest").Return(nil)
	mockDocker.On("CreateContainer", mock.Anything, mock.AnythingOfType("*workflow.ContainerConfig")).Return("", assert.AnError)

	ctx := t.Context()
	result, err := executor.Execute(ctx, step, runtime)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create container")
	mockDocker.AssertExpectations(t)
}

func TestShellStepExecutor_Execute_ContainerStartFailure(t *testing.T) {
	mockDocker := NewMockDockerClient()
	executor := CreateTestShellExecutor(mockDocker)

	step := CreateTestStep("start-fail-step", "Start Fail Step", "echo 'test'")
	runtime := CreateTestRuntime("/tmp/workspace")

	mockDocker.On("PullImage", mock.Anything, "ubuntu:latest").Return(nil)
	mockDocker.On("CreateContainer", mock.Anything, mock.AnythingOfType("*workflow.ContainerConfig")).Return("container-456", nil)
	mockDocker.On("StartContainer", mock.Anything, "container-456").Return(assert.AnError)
	mockDocker.On("StopContainer", mock.Anything, "container-456").Return(nil)
	mockDocker.On("RemoveContainer", mock.Anything, "container-456").Return(nil)

	ctx := t.Context()
	result, err := executor.Execute(ctx, step, runtime)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to start container")

	mockDocker.AssertExpectations(t)
	assert.Empty(t, runtime.Containers)
}

func TestShellStepExecutor_buildEnvironment(t *testing.T) {
	executor := &ShellStepExecutor{}

	step := &Step{
		ID:   "env-step",
		Name: "Environment Step",
		Run:  "env",
		Env: map[string]string{
			"STEP_VAR":   "step_value",
			"COMMON_VAR": "step_override",
		},
	}

	runtime := CreateTestRuntime("/tmp/workspace")
	runtime.JobContext.Job.Env = map[string]string{
		"JOB_VAR":    "job_value",
		"COMMON_VAR": "job_value",
	}

	env := executor.buildEnvironment(step, runtime)

	AssertEnvironmentContains(t, env, "JOB_VAR=job_value")
	AssertEnvironmentContains(t, env, "STEP_VAR=step_value")
	AssertEnvironmentContains(t, env, "COMMON_VAR=step_override") // step should override job
	AssertEnvironmentContains(t, env, "GITHUB_WORKSPACE=/github/workspace")
	AssertEnvironmentContains(t, env, "GITHUB_ACTOR=rehearse")
	AssertEnvironmentContains(t, env, "RUNNER_OS=Linux")
	AssertEnvironmentContains(t, env, "RUNNER_ARCH=X64")

	// Verify we have the expected number of environment variables
	// Job vars (2) + Step vars (2, with override) + GitHub defaults (4) = 7
	expectedVars := []string{
		"JOB_VAR=job_value",
		"STEP_VAR=step_value",
		"COMMON_VAR=step_override",
		"GITHUB_WORKSPACE=/github/workspace",
		"GITHUB_ACTOR=rehearse",
		"GITHUB_REPOSITORY=local/repo",
		"RUNNER_OS=Linux",
		"RUNNER_ARCH=X64",
	}

	assert.GreaterOrEqual(t, len(env), len(expectedVars))
}

func TestShellStepExecutor_buildEnvironment_NoJobContext(t *testing.T) {
	executor := &ShellStepExecutor{}

	step := &Step{
		ID:   "no-job-step",
		Name: "No Job Context Step",
		Run:  "echo 'test'",
		Env: map[string]string{
			"STEP_ONLY": "value",
		},
	}

	runtime := &Runtime{
		WorkingDir: "/tmp",
	}

	env := executor.buildEnvironment(step, runtime)

	AssertEnvironmentContains(t, env, "STEP_ONLY=value")
	AssertEnvironmentContains(t, env, "GITHUB_WORKSPACE=/github/workspace")
	AssertEnvironmentContains(t, env, "RUNNER_OS=Linux")
}

func TestShellStepExecutor_buildEnvironment_EmptyEnvironments(t *testing.T) {
	executor := &ShellStepExecutor{}

	step := &Step{
		ID:   "empty-env-step",
		Name: "Empty Environment Step",
		Run:  "echo 'test'",
	}

	runtime := CreateTestRuntime("/tmp/workspace")
	runtime.JobContext.Job.Env = nil

	env := executor.buildEnvironment(step, runtime)

	AssertEnvironmentContains(t, env, "GITHUB_WORKSPACE=/github/workspace")
	AssertEnvironmentContains(t, env, "GITHUB_ACTOR=rehearse")
	AssertEnvironmentContains(t, env, "RUNNER_OS=Linux")

	assert.GreaterOrEqual(t, len(env), 5)
}

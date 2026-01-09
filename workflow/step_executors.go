package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/telton/rehearse/internal/logger"
)

// ShellStepExecutor handles steps with 'run' commands.
type ShellStepExecutor struct {
	Docker   DockerClient
	renderer *RunRenderer
}

// CanExecute returns true if this step has a 'run' command.
func (e *ShellStepExecutor) CanExecute(step *Step) bool {
	return step.Run != ""
}

// Execute runs a shell command in a container.
func (e *ShellStepExecutor) Execute(ctx context.Context, step *Step, runtime *Runtime) (*ExecutionStepResult, error) {
	// Default to ubuntu-latest if no container specified
	image := "ubuntu:latest"
	if runtime.JobContext != nil && runtime.JobContext.Job.Container != nil {
		image = runtime.JobContext.Job.Container.Image
	}

	if e.renderer != nil {
		e.renderer.RenderDockerPull(image)
	}
	if err := e.Docker.PullImage(ctx, image); err != nil {
		return nil, fmt.Errorf("failed to pull image %s: %w", image, err)
	}

	evaluatedCommand := e.evaluateExpressions(step.Run, runtime)

	env := e.buildEnvironment(step, runtime)

	volumes := []VolumeMount{
		{
			Source: runtime.WorkingDir,
			Target: "/github/workspace",
			Type:   "bind",
		},
	}

	if runtime.TempDir != "" {
		envFile := runtime.TempDir + "/GITHUB_ENV"
		outputFile := runtime.TempDir + "/GITHUB_OUTPUT"

		if _, err := os.Stat(envFile); os.IsNotExist(err) {
			if err := os.WriteFile(envFile, []byte{}, 0600); err != nil {
				return nil, fmt.Errorf("failed to create GITHUB_ENV file: %w", err)
			}
		}
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			if err := os.WriteFile(outputFile, []byte{}, 0600); err != nil {
				return nil, fmt.Errorf("failed to create GITHUB_OUTPUT file: %w", err)
			}
		}

		volumes = append(volumes,
			VolumeMount{
				Source: envFile,
				Target: "/github/env/GITHUB_ENV",
				Type:   "bind",
			},
			VolumeMount{
				Source: outputFile,
				Target: "/github/env/GITHUB_OUTPUT",
				Type:   "bind",
			},
		)

		env = append(env,
			"GITHUB_ENV=/github/env/GITHUB_ENV",
			"GITHUB_OUTPUT=/github/env/GITHUB_OUTPUT",
		)
	}

	config := &ContainerConfig{
		Image:      image,
		Cmd:        []string{"sh", "-c", evaluatedCommand},
		Env:        env,
		WorkingDir: "/github/workspace",
		Volumes:    volumes,
	}

	containerID, err := e.Docker.CreateContainer(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	runtime.Containers[step.ID] = &ContainerInfo{
		ID:     containerID,
		Image:  image,
		Status: "created",
	}

	defer func() {
		if err := e.Docker.StopContainer(ctx, containerID); err != nil {
			logger.Warn("Failed to stop container", "container_id", containerID, "error", err)
		}
		if err := e.Docker.RemoveContainer(ctx, containerID); err != nil {
			logger.Warn("Failed to remove container", "container_id", containerID, "error", err)
		}
		delete(runtime.Containers, step.ID)
	}()

	if err := e.Docker.StartContainer(ctx, containerID); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	runtime.Containers[step.ID].Status = "running"

	var exitCode int
	var containerError error

	if dockerClient, ok := e.Docker.(*RealDockerClient); ok {
		exitCode, containerError = dockerClient.WaitForContainer(ctx, containerID)

		if logs, err := dockerClient.GetContainerLogs(ctx, containerID); err == nil && logs != "" {
			cleanLogs := strings.TrimSpace(logs)
			if cleanLogs != "" && e.renderer != nil {
				e.renderer.RenderContainerOutput(logs)
			}
		}
	} else {
		exitCode = 0
	}

	if containerError != nil {
		return nil, fmt.Errorf("container execution failed: %w", containerError)
	}

	return &ExecutionStepResult{
		Success:  exitCode == 0,
		ExitCode: exitCode,
		Outputs:  make(map[string]string),
	}, nil
}

// buildEnvironment creates environment variables for the step.
func (e *ShellStepExecutor) buildEnvironment(step *Step, runtime *Runtime) []string {
	var env []string

	if runtime.JobContext != nil && runtime.JobContext.Job.Env != nil {
		for k, v := range runtime.JobContext.Job.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	for k, v := range runtime.DynamicEnv {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	if step.Env != nil {
		for k, v := range step.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	env = append(env,
		"GITHUB_WORKSPACE=/github/workspace",
		"GITHUB_ACTOR=rehearse",
		"GITHUB_REPOSITORY=local/repo",
		"RUNNER_OS=Linux",
		"RUNNER_ARCH=X64",
	)

	return env
}

// evaluateExpressions evaluates GitHub Actions expressions in a string.
func (e *ShellStepExecutor) evaluateExpressions(input string, runtime *Runtime) string {
	result := input

	for {
		start := strings.Index(result, "${{")
		if start == -1 {
			break
		}

		end := strings.Index(result[start:], "}}")
		if end == -1 {
			break
		}
		end += start + 2

		expression := result[start+3 : end-2]
		expression = strings.TrimSpace(expression)

		value := e.evaluateExpression(expression, runtime)

		result = result[:start] + value + result[end:]
	}

	return result
}

// evaluateExpression evaluates a single GitHub Actions expression.
func (e *ShellStepExecutor) evaluateExpression(expression string, runtime *Runtime) string {
	if strings.HasPrefix(expression, "steps.") && strings.Contains(expression, ".outputs.") {
		parts := strings.Split(expression, ".")
		if len(parts) >= 4 && parts[0] == "steps" && parts[2] == "outputs" {
			stepID := parts[1]
			outputName := parts[3]

			if stepOutputs, exists := runtime.StepOutputs[stepID]; exists {
				if value, exists := stepOutputs[outputName]; exists {
					return value
				}
			}
		}
	}

	if strings.HasPrefix(expression, "env.") {
		envVar := expression[4:]
		if value, exists := runtime.DynamicEnv[envVar]; exists {
			return value
		}
	}

	return ""
}

// ActionStepExecutor handles steps with 'uses' actions.
type ActionStepExecutor struct {
	Docker DockerClient
	Git    ExecutorGitRepo
}

// CanExecute returns true if this step uses an action.
func (e *ActionStepExecutor) CanExecute(step *Step) bool {
	return step.Uses != ""
}

// Execute runs an action (local, repository, or docker).
func (e *ActionStepExecutor) Execute(ctx context.Context, step *Step, runtime *Runtime) (*ExecutionStepResult, error) {
	actionRef := step.Uses

	switch {
	case strings.HasPrefix(actionRef, "./"):
		return e.executeLocalAction(ctx, step, runtime, actionRef)
	case strings.HasPrefix(actionRef, "docker://"):
		return e.executeDockerAction(ctx, step, runtime, actionRef)
	case strings.Contains(actionRef, "/"):
		return e.executeRepositoryAction(ctx, step, runtime, actionRef)
	default:
		return nil, fmt.Errorf("unsupported action format: %s", actionRef)
	}
}

// executeLocalAction runs an action from the local filesystem.
func (e *ActionStepExecutor) executeLocalAction(ctx context.Context, step *Step, runtime *Runtime, actionPath string) (*ExecutionStepResult, error) {
	fullPath := filepath.Join(runtime.WorkingDir, actionPath)

	metadata, err := e.Git.GetActionMetadata(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load action metadata: %w", err)
	}

	return e.executeActionWithMetadata(ctx, step, runtime, metadata, fullPath)
}

// executeDockerAction runs a Docker-based action.
func (e *ActionStepExecutor) executeDockerAction(ctx context.Context, step *Step, runtime *Runtime, dockerRef string) (*ExecutionStepResult, error) {
	image := strings.TrimPrefix(dockerRef, "docker://")

	if err := e.Docker.PullImage(ctx, image); err != nil {
		return nil, fmt.Errorf("failed to pull image %s: %w", image, err)
	}

	env := e.buildActionEnvironment(step, runtime)

	config := &ContainerConfig{
		Image:      image,
		Env:        env,
		WorkingDir: "/github/workspace",
		Volumes: []VolumeMount{
			{
				Source: runtime.WorkingDir,
				Target: "/github/workspace",
				Type:   "bind",
			},
		},
	}

	containerID, err := e.Docker.CreateContainer(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	defer func() {
		if err := e.Docker.StopContainer(ctx, containerID); err != nil {
			logger.Warn("Failed to stop container", "container_id", containerID, "error", err)
		}
		if err := e.Docker.RemoveContainer(ctx, containerID); err != nil {
			logger.Warn("Failed to remove container", "container_id", containerID, "error", err)
		}
	}()

	if err := e.Docker.StartContainer(ctx, containerID); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	return &ExecutionStepResult{
		Success:  true,
		ExitCode: 0,
		Outputs:  make(map[string]string),
	}, nil
}

// executeRepositoryAction downloads and runs an action from a Git repository.
func (e *ActionStepExecutor) executeRepositoryAction(ctx context.Context, step *Step, runtime *Runtime, repoRef string) (*ExecutionStepResult, error) {
	// Parse repository reference (owner/repo@ref)
	parts := strings.Split(repoRef, "@")
	repo := parts[0]
	ref := "main"
	if len(parts) > 1 {
		ref = parts[1]
	}

	actionDir := filepath.Join("/tmp", "rehearse-actions", strings.ReplaceAll(repo, "/", "-"), ref)

	repoURL := fmt.Sprintf("https://github.com/%s", repo)
	if err := e.Git.CloneAction(ctx, repoURL, ref, actionDir); err != nil {
		return nil, fmt.Errorf("failed to clone action %s: %w", repoRef, err)
	}

	metadata, err := e.Git.GetActionMetadata(actionDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load action metadata for %s: %w", repoRef, err)
	}

	return e.executeActionWithMetadata(ctx, step, runtime, metadata, actionDir)
}

// executeActionWithMetadata executes an action using its metadata.
func (e *ActionStepExecutor) executeActionWithMetadata(ctx context.Context, step *Step, runtime *Runtime, metadata *ActionMetadata, actionPath string) (*ExecutionStepResult, error) {
	switch metadata.Runs.Using {
	case "docker":
		return e.executeDockerActionFromMetadata(ctx, step, runtime, metadata, actionPath)
	case "node16", "node20":
		return e.executeNodeAction(ctx, step, runtime, metadata, actionPath)
	case "composite":
		return e.executeCompositeAction(ctx, step, runtime, metadata, actionPath)
	default:
		return nil, fmt.Errorf("unsupported action type: %s", metadata.Runs.Using)
	}
}

// executeDockerActionFromMetadata runs a Docker action using metadata.
func (e *ActionStepExecutor) executeDockerActionFromMetadata(ctx context.Context, step *Step, runtime *Runtime, metadata *ActionMetadata, actionPath string) (*ExecutionStepResult, error) {
	image := metadata.Runs.Image

	if strings.HasPrefix(image, "Dockerfile") {
		return nil, fmt.Errorf("dockerfile-based actions not yet supported")
	}

	if err := e.Docker.PullImage(ctx, image); err != nil {
		return nil, fmt.Errorf("failed to pull image %s: %w", image, err)
	}

	env := e.buildActionEnvironment(step, runtime)

	if step.With != nil {
		for k, v := range step.With {
			envName := fmt.Sprintf("INPUT_%s", strings.ToUpper(strings.ReplaceAll(k, "-", "_")))
			env = append(env, fmt.Sprintf("%s=%s", envName, v))
		}
	}

	config := &ContainerConfig{
		Image:      image,
		Env:        env,
		WorkingDir: "/github/workspace",
		Volumes: []VolumeMount{
			{
				Source: runtime.WorkingDir,
				Target: "/github/workspace",
				Type:   "bind",
			},
		},
	}

	containerID, err := e.Docker.CreateContainer(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	defer func() {
		if err := e.Docker.StopContainer(ctx, containerID); err != nil {
			logger.Warn("Failed to stop container", "container_id", containerID, "error", err)
		}
		if err := e.Docker.RemoveContainer(ctx, containerID); err != nil {
			logger.Warn("Failed to remove container", "container_id", containerID, "error", err)
		}
	}()

	if err := e.Docker.StartContainer(ctx, containerID); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	return &ExecutionStepResult{
		Success:  true,
		ExitCode: 0,
		Outputs:  make(map[string]string),
	}, nil
}

// executeNodeAction runs a Node.js-based action.
func (e *ActionStepExecutor) executeNodeAction(ctx context.Context, step *Step, runtime *Runtime, metadata *ActionMetadata, actionPath string) (*ExecutionStepResult, error) {
	nodeImage := "node:16"
	if metadata.Runs.Using == "node20" {
		nodeImage = "node:20"
	}

	env := e.buildActionEnvironment(step, runtime)

	if step.With != nil {
		for k, v := range step.With {
			envName := fmt.Sprintf("INPUT_%s", strings.ToUpper(strings.ReplaceAll(k, "-", "_")))
			env = append(env, fmt.Sprintf("%s=%s", envName, v))
		}
	}

	mainFile := metadata.Runs.Main
	if mainFile == "" {
		mainFile = "index.js"
	}

	config := &ContainerConfig{
		Image:      nodeImage,
		Cmd:        []string{"node", mainFile},
		Env:        env,
		WorkingDir: "/action",
		Volumes: []VolumeMount{
			{
				Source: runtime.WorkingDir,
				Target: "/github/workspace",
				Type:   "bind",
			},
			{
				Source: actionPath,
				Target: "/action",
				Type:   "bind",
			},
		},
	}

	containerID, err := e.Docker.CreateContainer(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	defer func() {
		if err := e.Docker.StopContainer(ctx, containerID); err != nil {
			logger.Warn("Failed to stop container", "container_id", containerID, "error", err)
		}
		if err := e.Docker.RemoveContainer(ctx, containerID); err != nil {
			logger.Warn("Failed to remove container", "container_id", containerID, "error", err)
		}
	}()

	if err := e.Docker.StartContainer(ctx, containerID); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	return &ExecutionStepResult{
		Success:  true,
		ExitCode: 0,
		Outputs:  make(map[string]string),
	}, nil
}

// executeCompositeAction runs a composite action (action with multiple steps).
func (e *ActionStepExecutor) executeCompositeAction(ctx context.Context, step *Step, runtime *Runtime, metadata *ActionMetadata, actionPath string) (*ExecutionStepResult, error) {
	for _, compositeStep := range metadata.Runs.Steps {
		if compositeStep.Run != "" {
			shellExecutor := &ShellStepExecutor{Docker: e.Docker}
			if _, err := shellExecutor.Execute(ctx, &compositeStep, runtime); err != nil {
				return nil, fmt.Errorf("composite step failed: %w", err)
			}
		}
	}

	return &ExecutionStepResult{
		Success:  true,
		ExitCode: 0,
		Outputs:  make(map[string]string),
	}, nil
}

// buildActionEnvironment creates environment variables for actions.
func (e *ActionStepExecutor) buildActionEnvironment(step *Step, runtime *Runtime) []string {
	var env []string

	if runtime.JobContext != nil && runtime.JobContext.Job.Env != nil {
		for k, v := range runtime.JobContext.Job.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	if step.Env != nil {
		for k, v := range step.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	env = append(env,
		"GITHUB_WORKSPACE=/github/workspace",
		"GITHUB_ACTOR=rehearse",
		"GITHUB_REPOSITORY=local/repo",
		"RUNNER_OS=Linux",
		"RUNNER_ARCH=X64",
	)

	return env
}

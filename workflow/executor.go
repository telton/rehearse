package workflow

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// DockerClient manages container lifecycle operations.
type DockerClient interface {
	CreateContainer(ctx context.Context, config *ContainerConfig) (string, error)
	StartContainer(ctx context.Context, containerID string) error
	ExecInContainer(ctx context.Context, containerID string, cmd []string) (*ExecResult, error)
	StopContainer(ctx context.Context, containerID string) error
	RemoveContainer(ctx context.Context, containerID string) error
	PullImage(ctx context.Context, image string) error
	Close() error
}

// ExecutorGitRepo manages git operations for action resolution.
type ExecutorGitRepo interface {
	CloneAction(ctx context.Context, repo, ref, dest string) error
	GetActionMetadata(path string) (*ActionMetadata, error)
	GetCurrentBranch() (string, error)
	GetCurrentCommit() (string, error)
}

// StepExecutor handles execution of different step types.
type StepExecutor interface {
	Execute(ctx context.Context, step *Step, runtime *Runtime) (*ExecutionStepResult, error)
	CanExecute(step *Step) bool
}

// Executor orchestrates workflow execution.
type Executor struct {
	analyzer  *Analyzer
	docker    DockerClient
	git       ExecutorGitRepo
	runtime   *Runtime
	executors []StepExecutor
	renderer  *RunRenderer
}

// Runtime tracks the execution state.
type Runtime struct {
	WorkingDir  string
	Containers  map[string]*ContainerInfo
	Networks    map[string]*NetworkInfo
	Volumes     map[string]*VolumeInfo
	JobContext  *ExecutionJobContext
	StepContext *ExecutionStepContext
	DynamicEnv  map[string]string            // Environment variables set during execution
	StepOutputs map[string]map[string]string // step_id -> output_name -> value
	TempDir     string                       // Directory for GITHUB_ENV and GITHUB_OUTPUT files
}

// ContainerConfig holds container creation parameters.
type ContainerConfig struct {
	Image      string
	Cmd        []string
	Env        []string
	WorkingDir string
	Volumes    []VolumeMount
	Networks   []string
}

// ContainerInfo tracks running container details.
type ContainerInfo struct {
	ID       string
	Image    string
	Status   string
	Networks []string
}

// NetworkInfo tracks Docker network details.
type NetworkInfo struct {
	ID   string
	Name string
}

// VolumeInfo tracks Docker volume details.
type VolumeInfo struct {
	ID         string
	Name       string
	MountPoint string
}

// VolumeMount represents a volume mount configuration.
type VolumeMount struct {
	Source string
	Target string
	Type   string // bind, volume, tmpfs
}

// ExecResult contains command execution results.
type ExecResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// ExecutionStepResult contains step execution results.
type ExecutionStepResult struct {
	Success  bool
	ExitCode int
	Outputs  map[string]string
	Error    error
	Duration int64 // nanoseconds
}

// ActionMetadata represents action.yml/action.yaml content.
type ActionMetadata struct {
	Name        string                  `yaml:"name"`
	Description string                  `yaml:"description"`
	Inputs      map[string]ActionInput  `yaml:"inputs"`
	Outputs     map[string]ActionOutput `yaml:"outputs"`
	Runs        ActionRuns              `yaml:"runs"`
}

// ActionInput represents an action input definition.
type ActionInput struct {
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
	Default     string `yaml:"default"`
}

// ActionOutput represents an action output definition.
type ActionOutput struct {
	Description string `yaml:"description"`
}

// ActionRuns defines how an action executes.
type ActionRuns struct {
	Using string            `yaml:"using"` // docker, node16, node20, composite
	Image string            `yaml:"image"` // for docker actions
	Main  string            `yaml:"main"`  // for js actions
	Steps []Step            `yaml:"steps"` // for composite actions
	Env   map[string]string `yaml:"env"`
}

// ExecutionJobContext holds job-level execution context.
type ExecutionJobContext struct {
	Job       *Job
	Matrix    map[string]any
	Outputs   map[string]string
	Status    string // success, failure, cancelled, skipped
	StartTime int64
	EndTime   int64
}

// ExecutionStepContext holds step-level execution context.
type ExecutionStepContext struct {
	Step       *Step
	Outputs    map[string]string
	Outcome    string // success, failure, cancelled, skipped
	Conclusion string // success, failure, cancelled, skipped, neutral
}

// NewExecutor creates a new workflow executor.
func NewExecutor(analyzer *Analyzer, docker DockerClient, git ExecutorGitRepo) *Executor {
	return &Executor{
		analyzer: analyzer,
		docker:   docker,
		git:      git,
		runtime: &Runtime{
			Containers:  make(map[string]*ContainerInfo),
			Networks:    make(map[string]*NetworkInfo),
			Volumes:     make(map[string]*VolumeInfo),
			DynamicEnv:  make(map[string]string),
			StepOutputs: make(map[string]map[string]string),
		},
		executors: []StepExecutor{
			&ShellStepExecutor{Docker: docker, renderer: NewRunRenderer()},
			&ActionStepExecutor{Docker: docker, Git: git},
		},
		renderer: NewRunRenderer(),
	}
}

// Execute runs the workflow with the given context.
func (e *Executor) Execute(ctx context.Context, workflow *Workflow, triggerContext *Context) error {
	if err := e.setupTempDirectory(); err != nil {
		return fmt.Errorf("setting up temp directory: %w", err)
	}
	defer e.cleanupTempDirectory()

	analysis := e.analyzer.Analyze()
	if analysis == nil {
		return fmt.Errorf("workflow analysis failed")
	}

	for _, jobResult := range analysis.Jobs {
		if !jobResult.WouldRun {
			continue
		}

		job, exists := workflow.Jobs[jobResult.Name]
		if !exists {
			return fmt.Errorf("job %s not found in workflow", jobResult.Name)
		}

		if err := e.executeJob(ctx, &job, triggerContext); err != nil {
			return fmt.Errorf("job %s failed: %w", jobResult.Name, err)
		}
	}

	return nil
}

// executeJob runs a single job.
func (e *Executor) executeJob(ctx context.Context, job *Job, triggerContext *Context) error {
	e.renderer.RenderJobStart(job.Name)

	e.runtime.JobContext = &ExecutionJobContext{
		Job:       job,
		Outputs:   make(map[string]string),
		Status:    "in_progress",
		StartTime: getCurrentTime(),
	}

	defer func() {
		e.runtime.JobContext.EndTime = getCurrentTime()
		duration := e.runtime.JobContext.EndTime - e.runtime.JobContext.StartTime
		status := e.runtime.JobContext.Status

		if status == "success" {
			e.processJobOutputs(job)
		}

		if status == "success" {
			e.renderer.RenderJobSuccess(job.Name, duration)
		} else {
			e.renderer.RenderJobError(job.Name, duration)
		}
	}()

	for i, step := range job.Steps {
		e.renderer.RenderStepStart(i+1, len(job.Steps), step.Name)

		if err := e.executeStep(ctx, &step, triggerContext); err != nil {
			e.runtime.JobContext.Status = "failure"
			return fmt.Errorf("step %s failed: %w", step.Name, err)
		}

		e.renderer.RenderStepSuccess(step.Name)
	}

	e.runtime.JobContext.Status = "success"
	return nil
}

// executeStep runs a single step.
func (e *Executor) executeStep(ctx context.Context, step *Step, triggerContext *Context) error {
	e.runtime.StepContext = &ExecutionStepContext{
		Step:    step,
		Outputs: make(map[string]string),
	}

	for _, executor := range e.executors {
		if executor.CanExecute(step) {
			result, err := executor.Execute(ctx, step, e.runtime)
			if err != nil {
				e.runtime.StepContext.Outcome = "failure"
				e.runtime.StepContext.Conclusion = "failure"
				e.renderer.RenderStepError(step.Name, err)
				return err
			}

			if result.Success {
				e.runtime.StepContext.Outcome = "success"
				e.runtime.StepContext.Conclusion = "success"
			} else {
				e.runtime.StepContext.Outcome = "failure"
				e.runtime.StepContext.Conclusion = "failure"
				return fmt.Errorf("step failed with exit code %d", result.ExitCode)
			}

			for k, v := range result.Outputs {
				e.runtime.StepContext.Outputs[k] = v
			}

			if err := e.processStepOutputFiles(step.ID); err != nil {
				e.renderer.RenderWarning("failed to process output files: " + err.Error())
			}

			return nil
		}
	}

	return fmt.Errorf("no executor found for step: %s", step.Name)
}

// SetWorkingDirectory sets the working directory for workflow execution.
func (e *Executor) SetWorkingDirectory(workingDir string) {
	e.runtime.WorkingDir = workingDir
}

// setupTempDirectory creates a temporary directory for GitHub environment files.
func (e *Executor) setupTempDirectory() error {
	tempDir, err := os.MkdirTemp("", "rehearse-github-")
	if err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}

	e.runtime.TempDir = tempDir
	return nil
}

// cleanupTempDirectory removes the temporary directory.
func (e *Executor) cleanupTempDirectory() {
	if e.runtime.TempDir != "" {
		os.RemoveAll(e.runtime.TempDir)
		e.runtime.TempDir = ""
	}
}

// processStepOutputFiles processes GITHUB_ENV and GITHUB_OUTPUT files after step execution.
func (e *Executor) processStepOutputFiles(stepID string) error {
	if e.runtime.TempDir == "" {
		return nil
	}

	envFile := e.runtime.TempDir + "/GITHUB_ENV"
	if content, err := os.ReadFile(envFile); err == nil && len(content) > 0 {
		contentStr := strings.TrimSpace(string(content))
		contentStr = strings.ReplaceAll(contentStr, "\x00", "") // Remove null bytes

		if contentStr != "" {
			lines := strings.Split(contentStr, "\n")
			for _, line := range lines {
				if line = strings.TrimSpace(line); line != "" {
					if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
						key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
						if key != "" { // Only add non-empty keys
							e.runtime.DynamicEnv[key] = value
							e.renderer.RenderEnvironmentSet(key, value)
						}
					}
				}
			}
		}

		if err := os.WriteFile(envFile, []byte{}, 0600); err != nil {
			return fmt.Errorf("clear GITHUB_ENV file: %w", err)
		}
	}

	outputFile := e.runtime.TempDir + "/GITHUB_OUTPUT"
	if content, err := os.ReadFile(outputFile); err == nil && len(content) > 0 {
		if e.runtime.StepOutputs[stepID] == nil {
			e.runtime.StepOutputs[stepID] = make(map[string]string)
		}

		contentStr := strings.TrimSpace(string(content))
		contentStr = strings.ReplaceAll(contentStr, "\x00", "") // Remove null bytes

		if contentStr != "" {
			lines := strings.Split(contentStr, "\n")
			for _, line := range lines {
				if line = strings.TrimSpace(line); line != "" {
					if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
						key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
						if key != "" { // Only add non-empty keys
							e.runtime.StepOutputs[stepID][key] = value
							e.renderer.RenderOutputSet(stepID, key, value)
						}
					}
				}
			}
		}

		if err := os.WriteFile(outputFile, []byte{}, 0600); err != nil {
			return fmt.Errorf("clear GITHUB_OUTPUT file: %w", err)
		}
	}

	return nil
}

// processJobOutputs processes job outputs based on the job's outputs configuration.
func (e *Executor) processJobOutputs(job *Job) {
	if len(job.Outputs) == 0 {
		return
	}

	if e.runtime.JobContext == nil {
		return
	}

	e.renderer.RenderJobOutputsStart()
	for outputName, outputExpression := range job.Outputs {
		value := e.evaluateOutputExpression(outputExpression)

		e.runtime.JobContext.Outputs[outputName] = value
		e.renderer.RenderJobOutput(outputName, value)
	}
}

// evaluateOutputExpression evaluates a simple output expression.
// This is a basic implementation that handles common patterns like ${{ steps.stepid.outputs.outputname }}
func (e *Executor) evaluateOutputExpression(expression string) string {
	// Remove ${{ }} wrapper if present, handling various whitespace
	expr := strings.TrimSpace(expression)
	if strings.HasPrefix(expr, "${{") && strings.HasSuffix(expr, "}}") {
		// Extract content between ${{ and }}
		inner := expr[3 : len(expr)-2]
		// Clean up all types of whitespace (spaces, tabs, newlines)
		inner = strings.TrimSpace(inner)
		// Normalize internal whitespace - replace any whitespace sequences with single spaces
		parts := strings.Fields(inner)
		expr = strings.Join(parts, " ")
	}

	// Handle steps.stepid.outputs.outputname pattern
	if strings.HasPrefix(expr, "steps.") && strings.Contains(expr, ".outputs.") {
		parts := strings.Split(expr, ".")
		if len(parts) >= 4 && parts[0] == "steps" && parts[2] == "outputs" {
			stepID := parts[1]
			outputName := parts[3]

			if stepOutputs, exists := e.runtime.StepOutputs[stepID]; exists {
				if value, exists := stepOutputs[outputName]; exists {
					return value
				}
			}
		}
	}

	if strings.HasPrefix(expr, "env.") {
		envVar := expr[4:] // Remove "env." prefix
		if value, exists := e.runtime.DynamicEnv[envVar]; exists {
			return value
		}

		return ""
	}

	// If we get here, it's either a literal value or an unresolved expression
	// If it looks like an expression that we couldn't resolve, return empty string
	if strings.HasPrefix(expr, "steps.") || strings.HasPrefix(expr, "env.") {
		return ""
	}

	return expression
}

// getCurrentTime returns current unix timestamp in seconds.
func getCurrentTime() int64 {
	return time.Now().Unix()
}

package workflow

import (
	"fmt"
	"strings"

	"github.com/telton/rehearse/internal/logger"
	"github.com/telton/rehearse/ui"
)

// RunRenderer handles styled output for workflow execution
type RunRenderer struct{}

// NewRunRenderer creates a new run renderer
func NewRunRenderer() *RunRenderer {
	return &RunRenderer{}
}

// RenderWorkflowStart renders the initial workflow information
func (r *RunRenderer) RenderWorkflowStart(workflowName, workingDir, event, ref string) {
	logger.Debug("Rendering workflow start", "workflow", workflowName, "working_dir", workingDir, "event", event, "ref", ref)

	title := ui.NewHeader(workflowName).WithEmoji("*").WithMargin()
	fmt.Println(title.Render())

	workDir := ui.NewLabelValue("[DIR] Working directory:", workingDir)
	fmt.Println(workDir.Render())

	eventInfo := ui.NewLabelValue("[EVENT] Event:", event)
	fmt.Println(eventInfo.Render())

	if ref != "" {
		refInfo := ui.NewLabelValue("[REF] Ref:", ref)
		fmt.Println(refInfo.Render())
	}
	fmt.Println()
}

// RenderDockerCheck renders Docker availability check
func (r *RunRenderer) RenderDockerCheck() {
	status := ui.NewStatus("info", "Checking Docker availability...").WithIcon("[CHECK]")
	fmt.Println(status.Render())
}

// RenderDockerSuccess renders successful Docker connection
func (r *RunRenderer) RenderDockerSuccess() {
	status := ui.NewStatus("success", "Docker is available").WithIcon("[OK]")
	fmt.Println(status.Render())
}

// RenderDockerError renders Docker connection error
func (r *RunRenderer) RenderDockerError(err error) {
	warning := ui.NewStatus("warning", "Warning: "+err.Error()).WithIcon("[WARN]")
	fmt.Println(warning.Render())

	suggestion := ui.NewStatus("warning", "To run workflows locally, please install and start Docker").WithIcon("[TIP]")
	fmt.Println(suggestion.Render())

	link := ui.NewStatus("info", "Visit: https://docs.docker.com/get-docker/").WithIcon("   ")
	fmt.Println(link.Render())
}

// RenderDockerInit renders Docker client initialization
func (r *RunRenderer) RenderDockerInit() {
	status := ui.NewStatus("info", "Initializing Docker client...").WithIcon("[DOCKER]")
	fmt.Println(status.Render())
}

// RenderExecutionStart renders the start of workflow execution
func (r *RunRenderer) RenderExecutionStart() {
	status := ui.NewStatus("info", "Starting workflow execution...").WithIcon("[START]")
	fmt.Println(status.Render())
}

// RenderJobStart renders the start of a job
func (r *RunRenderer) RenderJobStart(jobName string) {
	logger.Debug("Rendering job start", "job", jobName)

	renderer := ui.NewWorkflowRenderer()
	header := renderer.RenderJobHeader("", jobName)
	fmt.Println("[RUN] " + header)
}

// RenderJobSuccess renders successful job completion
func (r *RunRenderer) RenderJobSuccess(jobName string, duration int64) {
	message := fmt.Sprintf("Job %s completed successfully in %ds", jobName, duration)
	status := ui.NewStatus("success", message).WithIcon("[OK]")
	fmt.Println(status.Render())
	fmt.Println()
}

// RenderJobError renders job failure
func (r *RunRenderer) RenderJobError(jobName string, duration int64) {
	message := fmt.Sprintf("Job %s failed after %ds", jobName, duration)
	status := ui.NewStatus("error", message).WithIcon("[FAIL]")
	fmt.Println(status.Render())
	fmt.Println()
}

// RenderStepStart renders the start of a step
func (r *RunRenderer) RenderStepStart(stepNum, totalSteps int, stepName string) {
	logger.Debug("Rendering step start", "step_num", stepNum, "total_steps", totalSteps, "step_name", stepName)

	message := fmt.Sprintf("Step %d/%d: %s", stepNum, totalSteps, stepName)
	status := ui.NewStatus("info", message).WithIcon("[STEP]")
	formatted := ui.WithMargin(ui.Muted, 2).Render(status.Render())
	fmt.Println(formatted)
}

// RenderStepSuccess renders successful step completion
func (r *RunRenderer) RenderStepSuccess(stepName string) {
	status := ui.NewStatus("success", stepName).WithIcon("[OK]")
	formatted := ui.WithMargin(ui.Muted, 2).Render(status.Render())
	fmt.Println(formatted)
}

// RenderStepError renders step failure
func (r *RunRenderer) RenderStepError(stepName string, err error) {
	message := fmt.Sprintf("%s - %v", stepName, err)
	status := ui.NewStatus("error", message).WithIcon("[FAIL]")
	formatted := ui.WithMargin(ui.Muted, 2).Render(status.Render())
	fmt.Println(formatted)
}

// RenderDockerPull renders Docker image pulling
func (r *RunRenderer) RenderDockerPull(image string) {
	renderer := ui.NewWorkflowRenderer()
	message := renderer.RenderDockerOperation("Pulling image", image)
	formatted := ui.WithMargin(ui.Muted, 4).Render(message)
	fmt.Println(formatted)
}

// RenderEnvironmentSet renders environment variable setting
func (r *RunRenderer) RenderEnvironmentSet(key, value string) {
	renderer := ui.NewWorkflowRenderer()
	message := renderer.RenderEnvironmentVar(key, value)
	status := ui.NewStatus("info", message).WithIcon("[ENV]")
	formatted := ui.WithMargin(ui.Muted, 4).Render(status.Render())
	fmt.Println(formatted)
}

// RenderOutputSet renders step output setting
func (r *RunRenderer) RenderOutputSet(stepID, key, value string) {
	message := fmt.Sprintf("Set output: %s.%s=%s", stepID, key, value)
	status := ui.NewStatus("info", message).WithIcon("[OUT]")
	formatted := ui.WithMargin(ui.Muted, 4).Render(status.Render())
	fmt.Println(formatted)
}

// RenderContainerOutput renders container output/logs
func (r *RunRenderer) RenderContainerOutput(logs string) {
	if logs == "" {
		return
	}

	outputHeader := ui.NewStatus("info", "Output:").WithIcon("[LOG]")
	formatted := ui.WithMargin(ui.Muted, 4).Render(outputHeader.Render())
	fmt.Println(formatted)

	// Clean up Docker log formatting and print with proper indentation
	cleanLogs := strings.TrimSpace(logs)
	for _, line := range strings.Split(cleanLogs, "\n") {
		// Skip Docker log stream headers (they start with special bytes)
		if len(line) > 8 {
			line = line[8:] // Remove Docker log header
		}
		if line != "" {
			renderer := ui.NewWorkflowRenderer()
			output := renderer.RenderOutput("  "+line, 6, false)
			fmt.Println(output)
		}
	}
}

// RenderJobOutputsStart renders the start of job output processing
func (r *RunRenderer) RenderJobOutputsStart() {
	status := ui.NewStatus("info", "Processing job outputs:").WithIcon("[STEP]")
	formatted := ui.WithMargin(ui.Muted, 4).Render(status.Render())
	fmt.Println(formatted)
}

// RenderJobOutput renders a single job output
func (r *RunRenderer) RenderJobOutput(name, value string) {
	message := fmt.Sprintf("%s = %s", name, value)
	renderer := ui.NewWorkflowRenderer()
	output := renderer.RenderOutput("  "+message, 6, false)
	fmt.Println(output)
}

// RenderWorkflowSuccess renders successful workflow completion
func (r *RunRenderer) RenderWorkflowSuccess() {
	status := ui.NewStatus("success", "Workflow execution completed successfully!").WithIcon("[OK]")
	fmt.Println(status.Render())
}

// RenderWorkflowError renders workflow execution error
func (r *RunRenderer) RenderWorkflowError(err error) {
	status := ui.NewStatus("error", "Workflow execution failed:").WithIcon("[FAIL]")
	fmt.Println(status.Render())

	errorDetails := ui.NewStatus("error", "   "+err.Error())
	fmt.Println(errorDetails.Render())
}

// RenderExecutionSummary renders a summary of the workflow execution
func (r *RunRenderer) RenderExecutionSummary(jobsRun, jobsFailed, stepsRun, stepsFailed int, totalDuration int64) {
	fmt.Println()

	renderer := ui.NewWorkflowRenderer()
	summary := renderer.RenderSummary(jobsRun, jobsRun-jobsFailed, jobsFailed, 0)
	fmt.Println(summary)

	if stepsFailed == 0 {
		stepStatus := ui.NewStatus("success", fmt.Sprintf("%d step(s) executed successfully", stepsRun)).WithIcon("[OK]")
		fmt.Println(ui.WithMargin(ui.Muted, 2).Render(stepStatus.Render()))
	} else {
		stepStatus := ui.NewStatus("error", fmt.Sprintf("%d step(s) executed, %d failed", stepsRun-stepsFailed, stepsFailed)).WithIcon("[FAIL]")
		fmt.Println(ui.WithMargin(ui.Muted, 2).Render(stepStatus.Render()))
	}

	timeInfo := ui.NewLabelValue("[TIME] Total time:", fmt.Sprintf("%ds", totalDuration)).WithIndent(2)
	fmt.Println(timeInfo.Render())
}

// RenderSeparator renders a visual separator
func (r *RunRenderer) RenderSeparator() {
	separator := ui.NewSeparator()
	fmt.Println(separator.Render())
}

// RenderWarning renders a general warning message
func (r *RunRenderer) RenderWarning(message string) {
	warning := ui.NewStatus("warning", "Warning: "+message).WithIcon("[WARN]")
	formatted := ui.WithMargin(ui.Muted, 4).Render(warning.Render())
	fmt.Println(formatted)
}

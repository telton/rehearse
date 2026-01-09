package ui

import (
	"fmt"
	"strings"
)

// WorkflowRenderer provides specialized rendering for workflow operations
type WorkflowRenderer struct {
	theme Theme
}

// NewWorkflowRenderer creates a workflow-specific renderer
func NewWorkflowRenderer() *WorkflowRenderer {
	return &WorkflowRenderer{theme: DefaultTheme()}
}

// RenderWorkflowHeader renders a workflow title with metadata
func (r *WorkflowRenderer) RenderWorkflowHeader(name, trigger string) string {
	header := NewHeader(fmt.Sprintf("Workflow: %s", name)).WithEmoji("🎭").WithMargin()
	triggerInfo := NewLabelValue("Trigger:", trigger).WithIndent(2)

	return header.Render() + "\n" + triggerInfo.Render()
}

// RenderContext renders GitHub context information
func (r *WorkflowRenderer) RenderContext(contextData map[string]string) string {
	header := NewHeader("Context:").WithMargin()
	var lines []string

	lines = append(lines, header.Render())

	for key, value := range contextData {
		lv := NewLabelValue(fmt.Sprintf("github.%-10s", key), value).WithIndent(2)
		lines = append(lines, lv.Render())
	}

	return strings.Join(lines, "\n")
}

// RenderJobHeader renders a job section header
func (r *WorkflowRenderer) RenderJobHeader(jobID, name string) string {
	title := jobID
	if name != "" && name != jobID {
		title = fmt.Sprintf("%s (%s)", jobID, name)
	}

	return WithColor(Bold, theme.Data).Render("🔧 Job: " + title)
}

// RenderStep renders a workflow step with status
func (r *WorkflowRenderer) RenderStep(name, status string, indent int) string {
	var icon string
	switch status {
	case "success", "completed":
		icon = "✓"
	case "error", "failed":
		icon = "✗"
	case "skipped":
		icon = "⊝"
	case "running":
		icon = "⟳"
	default:
		icon = "•"
	}

	stepText := fmt.Sprintf("%s %s", icon, name)
	style := StatusColor(status)
	if indent > 0 {
		style = WithMargin(style, indent)
	}

	return style.Render(stepText)
}

// RenderOutput renders command output with appropriate formatting
func (r *WorkflowRenderer) RenderOutput(text string, indent int, isError bool) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	var formatted []string

	style := WithColor(Muted, theme.Muted)
	if isError {
		style = WithColor(Muted, theme.Error)
	}

	for _, line := range lines {
		if indent > 0 {
			line = strings.Repeat(" ", indent) + line
		}
		formatted = append(formatted, style.Render(line))
	}

	return strings.Join(formatted, "\n")
}

// RenderSummary renders a workflow execution summary
func (r *WorkflowRenderer) RenderSummary(total, success, failed, skipped int) string {
	header := NewHeader("Summary").WithEmoji("📊").WithMargin()

	var summaryLines []string
	summaryLines = append(summaryLines, header.Render())

	if total > 0 {
		summaryLines = append(summaryLines,
			NewLabelValue("Total jobs:", fmt.Sprintf("%d", total)).WithIndent(2).Render())
	}
	if success > 0 {
		summaryLines = append(summaryLines,
			WithMargin(Success, 2).Render(fmt.Sprintf("✓ %d successful", success)))
	}
	if failed > 0 {
		summaryLines = append(summaryLines,
			WithMargin(Error, 2).Render(fmt.Sprintf("✗ %d failed", failed)))
	}
	if skipped > 0 {
		summaryLines = append(summaryLines,
			WithMargin(Warning, 2).Render(fmt.Sprintf("⊝ %d skipped", skipped)))
	}

	return strings.Join(summaryLines, "\n")
}

// RenderDockerOperation renders Docker-related operations
func (r *WorkflowRenderer) RenderDockerOperation(operation, image string) string {
	return WithColor(Bold, theme.Info).Render("🐳 " + operation + ": " + image)
}

// RenderEnvironmentVar renders environment variable assignments
func (r *WorkflowRenderer) RenderEnvironmentVar(key, value string) string {
	return WithColor(Muted, theme.Variable).Render(fmt.Sprintf("export %s=%s", key, value))
}

// RenderExpression renders GitHub Actions expressions
func (r *WorkflowRenderer) RenderExpression(expr, result string) string {
	exprText := WithColor(Code, theme.Special).Render(expr)
	resultText := WithColor(Value, theme.Data).Render(result)
	return fmt.Sprintf("%s → %s", exprText, resultText)
}

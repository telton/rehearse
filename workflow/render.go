package workflow

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	green  = lipgloss.Color("10")
	red    = lipgloss.Color("9")
	gray   = lipgloss.Color("8")
	pink   = lipgloss.Color("212")
	purple = lipgloss.Color("99")
	cyan   = lipgloss.Color("14")

	// Styles
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(purple)
	labelStyle   = lipgloss.NewStyle().Foreground(gray)
	valueStyle   = lipgloss.NewStyle().Foreground(cyan)
	passStyle    = lipgloss.NewStyle().Foreground(green)
	failStyle    = lipgloss.NewStyle().Foreground(red)
	skipStyle    = lipgloss.NewStyle().Foreground(gray)
	exprStyle    = lipgloss.NewStyle().Foreground(pink)
	boldStyle    = lipgloss.NewStyle().Bold(true)
	jobBoxStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(0, 1)
	summaryStyle = lipgloss.NewStyle().Bold(true).Foreground(cyan)
)

func Render(result *AnalysisResult) {
	fmt.Println(headerStyle.Render("Workflow: " + result.WorkflowName))
	fmt.Println(labelStyle.Render("Trigger: ") + valueStyle.Render(result.Trigger))
	fmt.Println()

	fmt.Println(headerStyle.Render("Context:"))
	fmt.Printf("  %s = %s\n", labelStyle.Render("github.ref       "), valueStyle.Render(result.Context.GitHub.Ref))
	fmt.Printf("  %s = %s\n", labelStyle.Render("github.event_name"), valueStyle.Render(result.Context.GitHub.EventName))
	fmt.Printf("  %s = %s\n", labelStyle.Render("github.sha       "), valueStyle.Render(truncateSHA(result.Context.GitHub.SHA)))
	fmt.Printf("  %s = %s\n", labelStyle.Render("github.actor     "), valueStyle.Render(result.Context.GitHub.Actor))
	fmt.Printf("  %s = %s\n", labelStyle.Render("github.repository"), valueStyle.Render(result.Context.GitHub.Repository))
	fmt.Println()

	willRun := 0
	skipped := 0

	for _, job := range result.Jobs {
		fmt.Println(renderJob(job))
		fmt.Println()

		if job.WouldRun {
			willRun++
		} else {
			skipped++
		}
	}

	summary := fmt.Sprintf("Summary: %d job(s) will run", willRun)
	if skipped > 0 {
		summary += fmt.Sprintf(", %d skipped", skipped)
	}
	fmt.Println(summaryStyle.Render(summary))
}

func renderJob(job JobResult) string {
	var b strings.Builder

	icon := passStyle.Render("[OK]")
	nameStyle := boldStyle
	if !job.WouldRun {
		icon = skipStyle.Render("[SKIP]")
		nameStyle = skipStyle.Bold(true)
	}

	header := fmt.Sprintf("%s Job: %s", icon, nameStyle.Render(job.Name))
	if !job.WouldRun {
		header += skipStyle.Render(" (SKIPPED)")
	}
	b.WriteString(header + "\n")

	b.WriteString(labelStyle.Render("runs-on: ") + job.RunsOn + "\n")

	if len(job.Needs) > 0 {
		b.WriteString(labelStyle.Render("needs: ") + "[" + strings.Join(job.Needs, ", ") + "]\n")
	}

	if job.Condition != nil {
		resultStr := passStyle.Render("TRUE")
		if !job.Condition.Value {
			resultStr = failStyle.Render("FALSE")
		}
		b.WriteString(fmt.Sprintf("%s %s -> %s\n", labelStyle.Render("if:"), exprStyle.Render(job.Condition.Expression), resultStr))
	}

	if len(job.Steps) > 0 {
		b.WriteString("\n")
		for _, step := range job.Steps {
			b.WriteString(renderStep(step, job.WouldRun) + "\n")
		}
	}

	content := strings.TrimSuffix(b.String(), "\n")

	return jobBoxStyle.Render(content)
}

func renderStep(step StepResult, jobWillRun bool) string {
	var icon string
	var nameStyle lipgloss.Style

	if !jobWillRun {
		icon = skipStyle.Render(".")
		nameStyle = skipStyle
	} else if step.WouldRun {
		icon = passStyle.Render("[OK]")
		nameStyle = lipgloss.NewStyle()
	} else {
		icon = skipStyle.Render("[SKIP]")
		nameStyle = skipStyle
	}

	line := fmt.Sprintf("  %s %s", icon, nameStyle.Render(step.Name))

	if step.Condition != nil {
		resultStr := passStyle.Render("TRUE")
		if !step.Condition.Value {
			resultStr = failStyle.Render("FALSE")
		}
		line += fmt.Sprintf("\n      %s %s → %s",
			labelStyle.Render("if:"),
			exprStyle.Render(step.Condition.Expression),
			resultStr)
	}

	return line
}

func truncateSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

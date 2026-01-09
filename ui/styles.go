package ui

import "github.com/charmbracelet/lipgloss"

// Base style configurations
var (
	// Text
	Bold      = lipgloss.NewStyle().Bold(true)
	Italic    = lipgloss.NewStyle().Italic(true)
	Underline = lipgloss.NewStyle().Underline(true)

	// Status using theme colors
	theme = DefaultTheme()

	Success = Bold.Bold(true).Foreground(theme.Success)
	Error   = Bold.Bold(true).Foreground(theme.Error)
	Warning = Bold.Bold(true).Foreground(theme.Warning)
	Info    = Bold.Bold(true).Foreground(theme.Info)
	Muted   = lipgloss.NewStyle().Foreground(theme.Muted)

	// Content
	Header = Bold.Bold(true).Foreground(theme.Emphasis)
	Label  = lipgloss.NewStyle().Foreground(theme.Muted)
	Value  = lipgloss.NewStyle().Foreground(theme.Data)
	Code   = lipgloss.NewStyle().Foreground(theme.Special)

	// Layout
	Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1)

	// Spacing helpers
	MarginLeft1  = lipgloss.NewStyle().MarginLeft(1)
	MarginLeft2  = lipgloss.NewStyle().MarginLeft(2)
	MarginLeft4  = lipgloss.NewStyle().MarginLeft(4)
	MarginLeft6  = lipgloss.NewStyle().MarginLeft(6)
	MarginBottom = lipgloss.NewStyle().MarginBottom(1)
)

// Semantic style builders
func StatusColor(status string) lipgloss.Style {
	switch status {
	case "success", "passed", "completed":
		return Success
	case "error", "failed":
		return Error
	case "warning", "skipped":
		return Warning
	case "info", "running":
		return Info
	default:
		return Muted
	}
}

// WithMargin adds left margin to any style
func WithMargin(style lipgloss.Style, left int) lipgloss.Style {
	return style.MarginLeft(left)
}

// WithColor applies color to any style
func WithColor(style lipgloss.Style, color lipgloss.Color) lipgloss.Style {
	return style.Foreground(color)
}

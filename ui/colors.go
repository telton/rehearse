package ui

import "github.com/charmbracelet/lipgloss"

var (
	Green  = lipgloss.Color("10")
	Red    = lipgloss.Color("9")
	Gray   = lipgloss.Color("8")
	Pink   = lipgloss.Color("212")
	Purple = lipgloss.Color("99")
	Cyan   = lipgloss.Color("14")
	Blue   = lipgloss.Color("12")
	Yellow = lipgloss.Color("11")
	Orange = lipgloss.Color("208")
)

// Theme provides semantic color access
type Theme struct {
	Success  lipgloss.Color
	Error    lipgloss.Color
	Warning  lipgloss.Color
	Info     lipgloss.Color
	Muted    lipgloss.Color
	Emphasis lipgloss.Color
	Data     lipgloss.Color
	Special  lipgloss.Color
	Variable lipgloss.Color
}

// DefaultTheme returns the standard rehearse color theme
func DefaultTheme() Theme {
	return Theme{
		Success:  Green,
		Error:    Red,
		Warning:  Yellow,
		Info:     Blue,
		Muted:    Gray,
		Emphasis: Purple,
		Data:     Cyan,
		Special:  Pink,
		Variable: Orange,
	}
}

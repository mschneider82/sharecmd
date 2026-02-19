package tui

import "github.com/charmbracelet/lipgloss"

var (
	Title   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	Success = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	Error   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
	Subtle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	Box     = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1)
)

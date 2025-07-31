package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Panel styles
	FocusedPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("69")).
				Padding(0, 1)

	UnfocusedPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("238")).
				Padding(0, 1)

	// Table styles
	HeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Bold(true)

	SelectedRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#01BE85")).
				Background(lipgloss.Color("#00432F"))

	RegularRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	DimRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	// Progress bar style
	ProgressStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF476F"))

	// Help text style
	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)
)

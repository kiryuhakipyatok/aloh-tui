package styles

import "github.com/charmbracelet/lipgloss"

var (
	BorderStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("35")).PaddingTop(1)

	TitleStyle   = lipgloss.NewStyle().Align(lipgloss.Center).Bold(true).Foreground(lipgloss.Color("63"))

	ContentStyle = lipgloss.NewStyle().Bold(true).MarginTop(1)

	FooterStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#CFCFCF")).Align(lipgloss.Center).MarginTop(1)
	
	HeaderStyle  = lipgloss.NewStyle().Align(lipgloss.Center).Bold(true).Foreground(lipgloss.Color("63"))
)
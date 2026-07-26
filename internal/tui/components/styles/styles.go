package styles

import "github.com/charmbracelet/lipgloss"

type UserColors struct {
	MainColor lipgloss.Color
	SubColor  lipgloss.Color
}

var (
	CDim  = lipgloss.Color("#75715E")
	CGray = lipgloss.Color("#5f5f5f")

	cAccent  = lipgloss.Color("#FD971F")
	cErr     = lipgloss.Color("#F92672")
	cSubtext = lipgloss.Color("#66D9EF")

	Black = "#1A1919"
	White = "#F8F8F2"

	CText = lipgloss.AdaptiveColor{Light: Black, Dark: White}

	WhitePulse = []string{
		"#444444",
		"#666666",
		"#888888",
		"#AAAAAA",
		"#CCCCCC",
		"#FFFFFF",
		"#CCCCCC",
		"#AAAAAA",
		"#888888",
		"#666666",
	}

	BlackPulse = []string{
		"#BBBBBB",
		"#999999",
		"#777777",
		"#555555",
		"#333333",
		"#000000",
		"#333333",
		"#555555",
		"#777777",
		"#999999",
	}

	PaddingLeftStyle      = lipgloss.NewStyle().PaddingLeft(2)
	CTextStyle            = lipgloss.NewStyle().Foreground(CText)
	CGrayStyle            = lipgloss.NewStyle().Foreground(CGray)
	CDimStyle             = lipgloss.NewStyle().Foreground(CDim)
	CSubTextStyle         = lipgloss.NewStyle().Foreground(cSubtext)
	CErrStyle             = lipgloss.NewStyle().Foreground(cErr)
	PaddingLeftCTextStyle = CTextStyle.PaddingLeft(2)
	PaddingLeftCGrayStyle = CGrayStyle.PaddingLeft(2)
	PaddingLeftCDimStyle  = CDimStyle.PaddingLeft(2)
	CErrPaddingStyle      = CErrStyle.PaddingRight(1)
	CErrBoldStyle         = CErrStyle.Bold(true)
	CGrayBold             = CGrayStyle.Bold(true)

	InactiveTabStyle = lipgloss.NewStyle().BorderForeground(CDim).BorderBottomForeground(CDim).Foreground(CDim).Align(lipgloss.Center)
	ActiveTabStyle   = lipgloss.NewStyle().BorderForeground(CDim).Bold(true).Align(lipgloss.Center)
	WindowStyle      = lipgloss.NewStyle().BorderForeground(CDim).Border(lipgloss.NormalBorder()).UnsetBorderTop()
	HeaderStyle      = CDimStyle.Bold(true)
)

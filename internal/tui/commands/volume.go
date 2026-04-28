package commands

import (
	"aloh-tui/internal/media/audio"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	FULL = iota
	MIC
)

type MuteMsg struct {
	Typee uint
	Res   bool
}

func MuteUnmuteCmd(ae audio.AudioEngine) tea.Cmd {
	return func() tea.Msg {
		res := ae.MuteUnmute()
		return MuteMsg{
			Typee: FULL,
			Res:   res,
		}
	}
}

func MuteUnmuteMicCmd(ae audio.AudioEngine) tea.Cmd {
	return func() tea.Msg {
		res := ae.MuteUnmuteMicro()
		return MuteMsg{
			Typee: MIC,
			Res:   res,
		}
	}
}
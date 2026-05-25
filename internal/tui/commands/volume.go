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
		msg := MuteMsg{}
		if ae != nil {
			res := ae.MuteUnmute()
			msg.Typee = FULL
			msg.Res = res
		}
		return msg
	}
}

func MuteUnmuteMicCmd(ae audio.AudioEngine) tea.Cmd {
	return func() tea.Msg {
		msg := MuteMsg{}
		if ae != nil {
			res := ae.MuteUnmuteMicro()
			msg.Typee = FULL
			msg.Res = res
		}
		return msg
	}
}

package commands

import (
	"aloh-tui/internal/media/audio"
	"aloh-tui/internal/networking"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	FULL = iota
	MIC
)

type MuteMsg struct {
	Typee uint
	Res   bool
	Err   error
}

func MuteUnmuteCmd(ae audio.AudioEngine, netw networking.Networking) tea.Cmd {
	return func() tea.Msg {
		msg := MuteMsg{}
		if ae != nil && netw != nil {
			res := ae.MuteUnmute()
			msg.Typee = FULL
			msg.Res = res

			e := networking.MuteFullEvent(msg.Res)
			if err := netw.NewEvent(e); err != nil {
				msg.Err = err
			}
		}

		return msg
	}
}

func MuteUnmuteMicCmd(ae audio.AudioEngine, netw networking.Networking) tea.Cmd {
	return func() tea.Msg {
		msg := MuteMsg{}
		if ae != nil && netw != nil {
			res := ae.MuteUnmuteMicro()
			msg.Typee = FULL
			msg.Res = res

			e := networking.MuteMicEvent(msg.Res)
			if err := netw.NewEvent(e); err != nil {
				msg.Err = err
			}
		}
		return msg
	}
}

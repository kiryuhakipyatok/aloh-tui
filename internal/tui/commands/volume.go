package commands

import (
	"aloh-tui/internal/entities/users"
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

func MuteUnmuteCmd(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := MuteMsg{}
		if user.Engines.AudioEngine != nil && user.Networking != nil {
			res := user.Engines.AudioEngine.MuteUnmute()
			msg.Typee = FULL
			msg.Res = res

			e, err := networking.MuteFullEvent(msg.Res)
			if err != nil {
				msg.Err = err
				return msg
			}
			if err := user.Networking.NewEvent(e); err != nil {
				msg.Err = err
				return msg
			}

			if err := user.MuteUnmuteFull(res); err != nil {
				msg.Err = err
			}

		}

		return msg
	}
}

func MuteUnmuteMicCmd(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := MuteMsg{}
		if user.Engines.AudioEngine != nil && user.Networking != nil {
			res := user.Engines.AudioEngine.MuteUnmuteMicro()
			msg.Typee = MIC
			msg.Res = res

			e, err := networking.MuteMicEvent(msg.Res)
			if err != nil {
				msg.Err = err
				return msg
			}
			if err := user.Networking.NewEvent(e); err != nil {
				msg.Err = err
			}
			if err := user.MuteUnmuteMic(res); err != nil {
				msg.Err = err
			}
		}
		return msg
	}
}

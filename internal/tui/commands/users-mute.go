package commands

import (
	"aloh-tui/internal/entities/users"

	tea "github.com/charmbracelet/bubbletea"
)

type MuteUnmuteUserMsg struct {
	Err error
}

func MuteUnmuteUserCmd(user *users.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := MuteUnmuteUserMsg{}
		res, err := user.Engines.AudioEngine.MuteUnmuteUser(nickname)
		if err != nil {
			msg.Err = err
			return msg
		}

		if err := user.MuteUnmuteUser(nickname, res); err != nil {
			msg.Err = err
		}

		return msg
	}

}

func SetupUserMuteCmd(user *users.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		var state bool
		us := user.GetUsersSetup(nickname)
		if us != nil {
			state = us.Muted
		}
		user.Engines.AudioEngine.SetMuteState(nickname, state)
		return MuteUnmuteUserMsg{}
	}
}

package commands

import (
	"aloh-tui/internal/entities/users"

	tea "github.com/charmbracelet/bubbletea"
)

type MuteUnmuteUserMsg struct {
	Err error
}

func MuteUnmuteUserCmd(user *users.User, iden users.Identity) tea.Cmd {
	return func() tea.Msg {
		msg := MuteUnmuteUserMsg{}
		if user.Engines.AudioEngine != nil {
			res, err := user.Engines.AudioEngine.MuteUnmuteUser(iden.ID)
			if err != nil {

			}

			if err := user.MuteUnmuteUser(iden.Nickname, res); err != nil {
				msg.Err = err
			}
		}
		return msg
	}

}

func SetupUserMuteCmd(user *users.User, iden users.Identity) tea.Cmd {
	return func() tea.Msg {
		msg := MuteUnmuteUserMsg{}
		var state bool
		if user.Engines.AudioEngine != nil {
			us, err := user.GetUsersSetup(iden.Nickname)
			if err != nil {
				msg.Err = err
				return msg
			}
			state = us.Muted
			user.Engines.AudioEngine.SetMuteState(iden.ID, state)
		}

		return msg
	}
}

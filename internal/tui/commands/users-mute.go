package commands

import (
	"aloh-tui/internal/entities/users"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type MuteUnmuteUserMsg struct {
	Err error
}

func MuteUnmuteUserCmd(user *users.User, nickname string, id uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		msg := MuteUnmuteUserMsg{}
		if user.Engines.AudioEngine != nil {
			res, err := user.Engines.AudioEngine.MuteUnmuteUser(id)
			if err != nil {

			}

			if err := user.MuteUnmuteUser(nickname, res); err != nil {
				msg.Err = err
			}
		}
		return msg
	}

}

func SetupUserMuteCmd(user *users.User, nickname string, id uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		msg := MuteUnmuteUserMsg{}
		var state bool
		if user.Engines.AudioEngine != nil {
			us, err := user.GetUsersSetup(nickname)
			if err != nil {
				msg.Err = err
				return msg
			}
			state = us.Muted
			user.Engines.AudioEngine.SetMuteState(id, state)
		}

		return msg
	}
}

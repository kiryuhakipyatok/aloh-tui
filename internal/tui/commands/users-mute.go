package commands

import (
	"aloh-tui/internal/entities"

	tea "github.com/charmbracelet/bubbletea"
)

type MuteUnmuteUserMsg struct {
	Err error
}

func MuteUnmuteUserCmd(user *entities.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := MuteUnmuteUserMsg{}
		res, err := user.Engines.AudioEngine.MuteUnmuteUser(nickname)
		if err != nil {
			msg.Err = err
			return msg
		}

		us, ok := user.Data.Setup.UsersSetup[nickname]
		if !ok {
			us = entities.UsersSetup{}
		}

		us.Muted = res

		user.Data.Setup.UsersSetup[nickname] = us

		if err := user.UpdateUserJSON(); err != nil {
			msg.Err = err
		}

		return msg
	}

}

func SetupUserMuteCmd(user *entities.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		var state bool
		us, ok := user.Data.Setup.UsersSetup[nickname]
		if ok {
			state = us.Muted
		}
		user.Engines.AudioEngine.SetMuteState(nickname, state)

		return MuteUnmuteUserMsg{}
	}
}

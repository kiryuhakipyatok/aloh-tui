package commands

import (
	"aloh-tui/internal/entities/users"

	tea "github.com/charmbracelet/bubbletea"
)

type UsersDenoiseMsg struct {
	Err error
}

func OnOffUsersHardDenoise(user *users.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := UsersDenoiseMsg{}
		if user.Engines.AudioEngine != nil {
			if err := user.Engines.AudioEngine.OnOffUsersHardDenoise(nickname); err != nil {
				msg.Err = err
			}
		}
		return msg
	}
}

func OnOffUsersSoftDenoise(user *users.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := UsersDenoiseMsg{}
		if user.Engines.AudioEngine != nil {
			if err := user.Engines.AudioEngine.OnOffUsersSoftDenoise(nickname); err != nil {
				msg.Err = err
			}
		}
		return msg
	}
}
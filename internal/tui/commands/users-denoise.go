package commands

import (
	"aloh-tui/internal/entities/users"

	tea "github.com/charmbracelet/bubbletea"
)

type UsersDenoiseMsg struct {
	Err error
}

func OnOffUsersHardDenoise(user *users.User, iden users.Identity, state bool) tea.Cmd {
	return func() tea.Msg {
		msg := UsersDenoiseMsg{}
		if user.Engines.AudioEngine != nil {
			if err := user.Engines.AudioEngine.OnOffUsersHardDenoise(iden.ID, state); err != nil {
				msg.Err = err
				return msg
			}
			if err := user.OnOffUsersHardDenoise(iden.Nickname, state); err != nil {
				msg.Err = err
			}
		}
		return msg
	}
}

func OnOffUsersSoftDenoise(user *users.User, iden users.Identity, state bool) tea.Cmd {
	return func() tea.Msg {
		msg := UsersDenoiseMsg{}
		if user.Engines.AudioEngine != nil {
			if err := user.Engines.AudioEngine.OnOffUsersSoftDenoise(iden.ID, state); err != nil {
				msg.Err = err
				return msg
			}
			if err := user.OnOffUsersSoftDenoise(iden.Nickname, state); err != nil {
				msg.Err = err
			}
		}
		return msg
	}
}

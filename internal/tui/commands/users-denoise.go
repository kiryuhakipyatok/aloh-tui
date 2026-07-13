package commands

import (
	"aloh-tui/internal/entities/users"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type UsersDenoiseMsg struct {
	Err error
}

func OnOffUsersHardDenoise(user *users.User, nickname string, id uuid.UUID, state bool) tea.Cmd {
	return func() tea.Msg {
		msg := UsersDenoiseMsg{}
		if user.Engines.AudioEngine != nil {
			if err := user.Engines.AudioEngine.OnOffUsersHardDenoise(id, state); err != nil {
				msg.Err = err
				return msg
			}
			if err := user.OnOffUsersHardDenoise(nickname, state); err != nil {
				msg.Err = err
			}
		}
		return msg
	}
}

func OnOffUsersSoftDenoise(user *users.User, nickname string, id uuid.UUID, state bool) tea.Cmd {
	return func() tea.Msg {
		msg := UsersDenoiseMsg{}
		if user.Engines.AudioEngine != nil {
			if err := user.Engines.AudioEngine.OnOffUsersSoftDenoise(id, state); err != nil {
				msg.Err = err
				return msg
			}
			if err := user.OnOffUsersSoftDenoise(nickname, state); err != nil {
				msg.Err = err
			}
		}
		return msg
	}
}

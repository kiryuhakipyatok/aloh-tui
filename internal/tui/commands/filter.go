package commands

import (
	"aloh-tui/internal/entities"

	tea "github.com/charmbracelet/bubbletea"
)

type OnOffFilterMsg struct {
	Err error
}

func OnOffFilterCmd(user *entities.User) tea.Cmd {
	return func() tea.Msg {
		msg := OnOffFilterMsg{}
		d := user.Engines.AudioEngine.OnOffFilter()
		user.Data.Setup.Filter = d

		if err := user.UpdateUserJSON(); err != nil {
			msg.Err = err
		}
		return msg
	}
}

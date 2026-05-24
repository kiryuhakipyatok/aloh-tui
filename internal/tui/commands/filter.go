package commands

import (
	"aloh-tui/internal/entities/users"

	tea "github.com/charmbracelet/bubbletea"
)

type OnOffFilterMsg struct {
	Err error
}

func OnOffFilterCmd(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := OnOffFilterMsg{}
		d := user.Engines.AudioEngine.OnOffFilter()
		if err := user.OnOffFilter(d); err != nil {
			msg.Err = err
		}
		return msg
	}
}

package commands

import (
	"aloh-tui/internal/entities/users"

	tea "github.com/charmbracelet/bubbletea"
)

type OnOffAECMsg struct {
	Err error
}

func OnOffAECCmd(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := OnOffAECMsg{}
		d := user.Engines.AudioEngine.OnOffAEC()
		if err := user.OnOffAEC(d); err != nil {
			msg.Err = err
		}
		return msg
	}
}

package commands

import (
	"aloh-tui/internal/entities"

	tea "github.com/charmbracelet/bubbletea"
)

type OnOffAECMsg struct {
	Err error
}

func OnOffAECCmd(user *entities.User) tea.Cmd {
	return func() tea.Msg {
		msg := OnOffAECMsg{}
		d := user.Engines.AudioEngine.OnOffAEC()
		user.Data.Setup.AEC = d

		if err := user.UpdateUserJSON(); err != nil {
			msg.Err = err
		}
		return msg
	}
}

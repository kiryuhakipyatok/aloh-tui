package commands

import (
	"aloh-tui/internal/entities"

	tea "github.com/charmbracelet/bubbletea"
)

type OnOffDenoiceMsg struct {
	Err error
}

func OnOffHardDenoiceCmd(user *entities.User) tea.Cmd {
	return func() tea.Msg {
		msg := OnOffDenoiceMsg{}
		h := user.Engines.AudioEngine.OnOffHardDenoice()
		user.Data.Setup.HardDenoise = h

		if err := user.UpdateUserJSON(); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func OnOffSoftDenoiceCmd(user *entities.User) tea.Cmd {
	return func() tea.Msg {
		msg := OnOffDenoiceMsg{}
		s := user.Engines.AudioEngine.OnOffSoftDenoice()
		user.Data.Setup.SoftDenoise = s

		if err := user.UpdateUserJSON(); err != nil {
			msg.Err = err
		}
		return msg
	}
}

package commands

import (
	"aloh-tui/internal/entities/users"

	tea "github.com/charmbracelet/bubbletea"
)

type OnOffDenoiceMsg struct {
	Err error
}

func OnOffHardDenoiceCmd(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := OnOffDenoiceMsg{}
		h := user.Engines.AudioEngine.OnOffHardDenoice()
		if err := user.OnOffHardDenoice(h); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func OnOffSoftDenoiceCmd(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := OnOffDenoiceMsg{}
		s := user.Engines.AudioEngine.OnOffSoftDenoice()
		if err := user.OnOffSoftDenoice(s); err != nil {
			msg.Err = err
		}
		return msg
	}
}

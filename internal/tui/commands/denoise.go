package commands

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/networking"

	tea "github.com/charmbracelet/bubbletea"
)

type OnOffDenoiceMsg struct {
	Err error
}

func OnOffHardDenoiceCmd(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := OnOffDenoiceMsg{}
		if user.Engines.AudioEngine != nil && user.Networking != nil {
			h := user.Engines.AudioEngine.OnOffHardDenoice()
			e, err := networking.HardDenoiseEvent(h)
			if err != nil {
				msg.Err = err
				return msg
			}
			if err := user.Networking.NewEvent(e); err != nil {
				msg.Err = err
				return msg
			}
			if err := user.OnOffHardDenoice(h); err != nil {
				msg.Err = err
			}
		}

		return msg
	}
}

func OnOffSoftDenoiceCmd(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := OnOffDenoiceMsg{}
		if user.Engines.AudioEngine != nil {
			s := user.Engines.AudioEngine.OnOffSoftDenoice()
			e, err := networking.SoftDenoiseEvent(s)
			if err != nil {
				msg.Err = err
				return msg
			}
			if err := user.Networking.NewEvent(e); err != nil {
				msg.Err = err
				return msg
			}
			if err := user.OnOffSoftDenoice(s); err != nil {
				msg.Err = err
			}
		}
		return msg
	}
}

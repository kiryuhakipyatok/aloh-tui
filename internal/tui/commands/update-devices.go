package commands

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/tui/components/lists"

	tea "github.com/charmbracelet/bubbletea"
)

type UpdateDevicesMsg struct {
	Typee uint
	Err   error
}

func UpdateMicrophonesCmd(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := UpdateDevicesMsg{
			Typee: lists.MICROPHONE,
		}
		if user.Engines.AudioEngine != nil {
			if err := user.Engines.AudioEngine.UpdateMicrophones(); err != nil {
				msg.Err = err
			}
		}

		return msg
	}
}

func UpdateHeadphonesCmd(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := UpdateDevicesMsg{
			Typee: lists.HEADPHONES,
		}
		if user.Engines.AudioEngine != nil {
			if err := user.Engines.AudioEngine.UpdateHeadphones(); err != nil {
				msg.Err = err
			}
		}

		return msg
	}
}

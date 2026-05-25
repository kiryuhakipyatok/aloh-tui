package commands

import (
	"aloh-tui/internal/entities/users"

	tea "github.com/charmbracelet/bubbletea"
)

type UpdateMicrophonesMsg struct {
	Err error
}

func UpdateMicrophonesCmd(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := UpdateMicrophonesMsg{}
		if user.Engines.AudioEngine != nil {
			if err := user.Engines.AudioEngine.UpdateMicrophones(); err != nil {
				msg.Err = err
			}
		}

		return msg
	}
}

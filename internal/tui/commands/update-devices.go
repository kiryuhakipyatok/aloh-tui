package commands

import (
	"aloh-tui/internal/entities"

	tea "github.com/charmbracelet/bubbletea"
)

type UpdateMicrophonesMsg struct {
	Err error
}

func UpdateMicrophonesCmd(user *entities.User) tea.Cmd {
	return func() tea.Msg {
		msg := UpdateMicrophonesMsg{}
		if err := user.Engines.AudioEngine.UpdateMicrophones(); err != nil {
			msg.Err = err
		}
		return msg
	}
}

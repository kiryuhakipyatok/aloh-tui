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
		if err := user.Engines.AudioEngine.UpdateMicrophones(); err != nil {
			return UpdateMicrophonesMsg{err}
		}
		return UpdateMicrophonesMsg{nil}
	}
}
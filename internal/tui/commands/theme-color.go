package commands

import (
	"aloh-tui/internal/entities"
	"encoding/json"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type ChangeThemeMsg struct {
	Err error
}

func ChangeThemeCmd(user *entities.User, newThemeColor string) tea.Cmd {
	return func() tea.Msg {
		user.Data.Setup.ThemeColor = newThemeColor

		userData, err := json.Marshal(user.Data)
		if err != nil {
			return ChangeThemeMsg{Err: err}
		}

		if err := os.WriteFile(user.Paths.DataFilePath, userData, 0644); err != nil {
			return ChangeThemeMsg{Err: err}
		}
		return ChangeThemeMsg{
			Err: nil,
		}
	}
}

package commands

import (
	"aloh-tui/internal/entities"

	tea "github.com/charmbracelet/bubbletea"
)

type ChangeThemeMsg struct {
	Err error
}

func ChangeThemeCmd(user *entities.User, newThemeColor string) tea.Cmd {
	return func() tea.Msg {
		msg := ChangeThemeMsg{}
		user.Data.Setup.ThemeColor = newThemeColor

		if err := user.UpdateUserJSON(); err != nil {
			msg.Err = err
		}
		return msg
	}
}

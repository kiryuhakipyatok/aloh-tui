package commands

import (
	"aloh-tui/internal/entities"

	tea "github.com/charmbracelet/bubbletea"
)

type ThemeColorMsg struct {
	Err error
}

type BFTagMsg struct {
	Err error
}

func ChangeThemeColorCmd(user *entities.User, newThemeColor string) tea.Cmd {
	return func() tea.Msg {
		msg := ThemeColorMsg{}
		user.Data.Setup.ThemeColor = newThemeColor

		if err := user.UpdateUserJSON(); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func ChangeBFTagCmd(user *entities.User, newBFTag string) tea.Cmd {
	return func() tea.Msg {
		msg := BFTagMsg{}
		user.Data.Setup.BestFriendTag = newBFTag

		if err := user.UpdateUserJSON(); err != nil {
			msg.Err = err
		}
		return msg
	}
}

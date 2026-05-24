package commands

import (
	"aloh-tui/internal/entities/users"

	tea "github.com/charmbracelet/bubbletea"
)

type ThemeColorMsg struct {
	Err error
}

type BFTagMsg struct {
	Err error
}

type NotificaionSignMsg struct {
	Err error
}

func ChangeThemeColorCmd(user *users.User, newThemeColor string) tea.Cmd {
	return func() tea.Msg {
		msg := ThemeColorMsg{}
		if err := user.ChangeThemeColor(newThemeColor); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func ChangeBFTagCmd(user *users.User, newBFTag string) tea.Cmd {
	return func() tea.Msg {
		msg := BFTagMsg{}
		if err := user.ChangeBFTag(newBFTag); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func ChangeNotificationSignCmd(user *users.User, newNotifySign string) tea.Cmd {
	return func() tea.Msg {
		msg := NotificaionSignMsg{}
		if err := user.ChangeNotificationSign(newNotifySign); err != nil {
			msg.Err = err
		}
		return msg
	}
}

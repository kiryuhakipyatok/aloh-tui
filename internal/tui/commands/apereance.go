package commands

import (
	"aloh-tui/internal/entities/users"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	COLOR = iota
	BFTAG
	N_TAG
	B_TAG
	S_TIME
	S_DATE
	S_WEATHER
	S_ZONE
)

type AppereanceMsg struct {
	Typee uint
	Err   error
}

func ChangeThemeColorCmd(user *users.User, newThemeColor string) tea.Cmd {
	return func() tea.Msg {
		msg := AppereanceMsg{
			Typee: COLOR,
		}
		if err := user.ChangeThemeColor(newThemeColor); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func ChangeBFTagCmd(user *users.User, newBFTag string) tea.Cmd {
	return func() tea.Msg {
		msg := AppereanceMsg{
			Typee: BFTAG,
		}
		if err := user.ChangeBFTag(newBFTag); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func ChangeNotificationTagCmd(user *users.User, newNotifyTag string) tea.Cmd {
	return func() tea.Msg {
		msg := AppereanceMsg{
			Typee: N_TAG,
		}
		if err := user.ChangeNotificationTag(newNotifyTag); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func ChangeBanTagCmd(user *users.User, newBanTag string) tea.Cmd {
	return func() tea.Msg {
		msg := AppereanceMsg{
			Typee: B_TAG,
		}
		if err := user.ChangeBanTag(newBanTag); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func OnOffShowTime(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := AppereanceMsg{
			Typee: S_TIME,
		}
		if err := user.OnOffShowTime(); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func OnOffShowDate(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := AppereanceMsg{
			Typee: S_DATE,
		}
		if err := user.OnOffShowDate(); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func OnOffShowZone(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := AppereanceMsg{
			Typee: S_ZONE,
		}
		if err := user.OnOffShowZone(); err != nil {
			msg.Err = err
		}
		return msg
	}
}

package commands

import (
	"aloh-tui/internal/entities/users"
	"context"
	"time"

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
	TAGLINE
)

type AppereanceMsg struct {
	Typee uint
	Res   bool
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

func ChangeTaglineCmd(user *users.User, newTagline string) tea.Cmd {
	return func() tea.Msg {
		msg := AppereanceMsg{
			Typee: TAGLINE,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := user.SSHClient.SetTagline(ctx, newTagline); err != nil {
			msg.Err = err
			return msg
		}
		if err := user.ChangeTagline(newTagline); err != nil {
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
		res, err := user.OnOffShowTime()
		if err != nil {
			msg.Err = err
		}
		msg.Res = res
		return msg
	}
}

func OnOffShowDate(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := AppereanceMsg{
			Typee: S_DATE,
		}
		res, err := user.OnOffShowDate()
		if err != nil {
			msg.Err = err
		}
		msg.Res = res
		return msg
	}
}

func OnOffShowZone(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := AppereanceMsg{
			Typee: S_ZONE,
		}
		res, err := user.OnOffShowZone()
		if err != nil {
			msg.Err = err
		}
		msg.Res = res
		return msg
	}
}

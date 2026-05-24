package commands

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/media/audio"
	"aloh-tui/internal/notifications"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type NotificationMessage struct {
	Err error
}

func NotifyCmd(nickname, msg string) tea.Cmd {
	return func() tea.Msg {
		cmsg := NotificationMessage{}
		t := time.Now().Format("15:04:05")
		if err := notifications.Notify(t, nickname, msg); err != nil {
			cmsg.Err = err
		}

		return cmsg
	}
}

func PlayNotificationCmd(ae audio.AudioEngine) tea.Cmd {
	return func() tea.Msg {
		ae.PlayNotification()
		return NotificationMessage{}
	}
}

func OnOffAudioNotifications(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := NotificationMessage{}
		if err := user.OnOffAudioNotification(); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func OnOffDesktopNotifications(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := NotificationMessage{}
		if err := user.OnOffDesktopNotification(); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func OnOffAppNotifications(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := NotificationMessage{}
		if err := user.OnOffAppNotification(); err != nil {
			msg.Err = err
		}
		return msg
	}
}

package commands

import (
	"aloh-tui/internal/entities"
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

func OnOffAudioNotifications(user *entities.User) tea.Cmd {
	return func() tea.Msg {
		cmsg := NotificationMessage{}
		user.Data.Setup.AudioNotifications = !user.Data.Setup.AudioNotifications

		if err := user.UpdateUserJSON(); err != nil {
			cmsg.Err = err
		}

		return cmsg
	}
}

func OnOffDesktopNotifications(user *entities.User) tea.Cmd {
	return func() tea.Msg {
		cmsg := NotificationMessage{}
		user.Data.Setup.DesktopNotifications = !user.Data.Setup.DesktopNotifications

		if err := user.UpdateUserJSON(); err != nil {
			cmsg.Err = err
		}

		return cmsg
	}
}

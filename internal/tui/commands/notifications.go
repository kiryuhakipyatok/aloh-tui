package commands

import (
	"aloh-tui/internal/media/audio"
	"aloh-tui/internal/notifications"

	tea "github.com/charmbracelet/bubbletea"
)

type NotificationMessage struct {
	Err error
}

func NotifyCmd(time, nickname, msg string) tea.Cmd {
	return func() tea.Msg {
		if err := notifications.Notify(time, nickname, msg); err != nil {
			return NotificationMessage{Err: err}
		}

		return NotificationMessage{Err: nil}
	}
}

func PlayNotificationCmd(ae audio.AudioEngine) tea.Cmd {
	return func() tea.Msg {
		ae.PlayNotification()
		return NotificationMessage{}
	}
}

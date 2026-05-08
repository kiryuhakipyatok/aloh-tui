package commands

import (
	"aloh-tui/internal/entities"
	"aloh-tui/internal/media/audio"
	"aloh-tui/internal/notifications"
	"encoding/json"
	"os"

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

func OnOffAudioNotifications(user *entities.User) tea.Cmd {
	return func() tea.Msg {
		user.Data.Setup.AudioNotifications = !user.Data.Setup.AudioNotifications

		userData, err := json.Marshal(user.Data)
		if err != nil {
			return NotificationMessage{Err: err}
		}

		if err := os.WriteFile(user.Paths.DataFilePath, userData, 0644); err != nil {
			return NotificationMessage{Err: err}
		}

		return NotificationMessage{Err: nil}
	}
}

func OnOffDesktopNotifications(user *entities.User) tea.Cmd {
	return func() tea.Msg {
		user.Data.Setup.DesktopNotifications = !user.Data.Setup.DesktopNotifications

		userData, err := json.Marshal(user.Data)
		if err != nil {
			return NotificationMessage{Err: err}
		}

		if err := os.WriteFile(user.Paths.DataFilePath, userData, 0644); err != nil {
			return NotificationMessage{Err: err}
		}

		return NotificationMessage{Err: nil}
	}
}

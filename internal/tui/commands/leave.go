package commands

import (
	"aloh-tui/internal/media/audio"
	"aloh-tui/internal/networking"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type LeaveMsg struct {
	Time string
	Err  error
}

func LeaveCmd(netw networking.Networking, ae audio.AudioEngine) tea.Cmd {
	return func() tea.Msg {
		t := time.Now().Format("15:04:05")
		msg := LeaveMsg{
			Time: t,
		}

		if err := netw.DisconnectFromUsers(); err != nil {
			msg.Err = err
		}
		ae.PlayNotification()
		return msg
	}
}

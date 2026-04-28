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
		if Err := netw.DisconnectFromUsers(); Err != nil {
			return LeaveMsg{Time: t, Err: Err}
		}
		ae.PlayNotification()
		return LeaveMsg{Time: t}
	}
}

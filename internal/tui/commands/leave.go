package commands

import (
	"aloh-tui/internal/networking"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type LeaveMsg struct {
	Time string
	Err  error
}

func LeaveCmd(netw networking.Networking) tea.Cmd {
	return func() tea.Msg {
		t := time.Now().Format("15:04:05")
		msg := LeaveMsg{
			Time: t,
		}
		if netw != nil {
			if err := netw.DisconnectFromAllUsers(); err != nil {
				msg.Err = err
			}
		}

		return msg
	}
}
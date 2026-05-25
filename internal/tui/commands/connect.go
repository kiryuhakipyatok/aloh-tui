package commands

import (
	"aloh-tui/internal/networking"

	tea "github.com/charmbracelet/bubbletea"
)

type ConnectToUserMsg struct {
	Err error
}

func ConnectToUserCmd(netw networking.Networking, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := ConnectToUserMsg{}
		if netw != nil {
			if err := netw.ConnectToUser(nickname); err != nil {
				msg.Err = err
			}
		}
		return msg
	}
}

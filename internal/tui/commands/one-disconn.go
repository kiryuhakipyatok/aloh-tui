package commands

import (
	"aloh-tui/internal/networking"

	tea "github.com/charmbracelet/bubbletea"
)

type SoloDisconn struct {
	Nickname string
	Err      error
}

func DisconnFromOne(netw networking.Networking, id string) tea.Cmd {
	return func() tea.Msg {
		msg := SoloDisconn{
			Nickname: id,
		}
		if netw != nil {
			if err := netw.DisconnectFromUser(id); err != nil {
				msg.Err = err
			}
		}
		return msg
	}
}

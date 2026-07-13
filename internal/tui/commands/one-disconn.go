package commands

import (
	"aloh-tui/internal/networking"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type SoloDisconn struct {
	Nickname string
	Err      error
}

func DisconnFromOne(netw networking.Networking, nickname string, id uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		msg := SoloDisconn{
			Nickname: nickname,
		}
		if netw != nil {
			if err := netw.DisconnectFromUser(id); err != nil {
				msg.Err = err
			}
		}
		return msg
	}
}

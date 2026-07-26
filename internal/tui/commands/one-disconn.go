package commands

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/networking"

	tea "github.com/charmbracelet/bubbletea"
)

type SoloDisconn struct {
	Identity users.Identity
	Err      error
}

func DisconnFromOne(netw networking.Networking, iden users.Identity) tea.Cmd {
	return func() tea.Msg {
		msg := SoloDisconn{
			Identity: iden,
		}
		if netw != nil {
			if err := netw.DisconnectFromUser(iden.ID); err != nil {
				msg.Err = err
			}
		}
		return msg
	}
}

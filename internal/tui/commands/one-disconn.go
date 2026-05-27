package commands

import (
	"aloh-tui/internal/networking"

	tea "github.com/charmbracelet/bubbletea"
)

type SoloDisconn struct {
	Err error
}

func DsiconnFromOne(netw networking.Networking, id string) tea.Cmd {
	return func() tea.Msg {
		msg := SoloDisconn{}
		if netw != nil {
			if err := netw.DisconnectFromUser(id); err != nil {
				msg.Err = err
			}
		}
		return msg
	}
}

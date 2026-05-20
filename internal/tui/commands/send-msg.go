package commands

import (
	"aloh-tui/internal/networking"

	tea "github.com/charmbracelet/bubbletea"
)

type SendInChatMsg struct {
	Err error
}

func SendInChatCmd(netw networking.Networking, msg []byte) tea.Cmd {
	return func() tea.Msg {
		cmsg := SendInChatMsg{}
		if err := netw.SendMessageInChat(msg); err != nil {
			cmsg.Err = err
		}
		return cmsg
	}
}

package commands

import (
	"aloh-tui/internal/networking"

	tea "github.com/charmbracelet/bubbletea"
)

type SendInChatMsg struct {
	Err error
}

func SendInChatCmd(netw networking.Networking, msg string) tea.Cmd {
	return func() tea.Msg {
		if err := netw.SendMessageInChat(msg); err != nil {
			return SendInChatMsg{err}
		}
		return SendInChatMsg{}
	}
}

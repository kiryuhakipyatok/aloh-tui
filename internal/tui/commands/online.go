package commands

import (
	"aloh-tui/internal/networking"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
)

type OnlineMsg struct {
	Online []string
	Err    error
}

func FetchOnlineCmd(netw networking.Networking, nickname string) tea.Cmd {
	return func() tea.Msg {
		onlineMsg := OnlineMsg{}
		Online, Err := netw.FetchCurrentOnline()
		if Err != nil {
			onlineMsg.Err = Err
		}

		onlineMsg.Online = slices.DeleteFunc(Online, func(v string) bool {
			return v == nickname
		})

		return onlineMsg
	}
}

package commands

import (
	"aloh-tui/internal/sshclient"

	tea "github.com/charmbracelet/bubbletea"
	alohnetwork "github.com/kiryuhakipyatok/aloh-networking"
)

func WaitForSSHEventMessageCmd(sub chan sshclient.Event) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}

type NetworkEventMsg struct {
	Nickname string
	Event    alohnetwork.Event
}

func WaitForNetworkEventMessageCmd(sub chan NetworkEventMsg) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}

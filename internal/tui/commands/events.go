package commands

import (
	"aloh-tui/internal/sshclient"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	alohnetwork "github.com/kiryuhakipyatok/aloh-networking"
)

func WaitForSSHEventMessageCmd(sub chan sshclient.Event) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}

type NetworkEventMsg struct {
	Id    uuid.UUID
	Event alohnetwork.Event
}

func WaitForNetworkEventMessageCmd(sub chan NetworkEventMsg) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}

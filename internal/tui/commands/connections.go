package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type PeerConnectedMsg struct {
	Id   uuid.UUID
	Time string
}

func WaitForPeerConnectionCmd(sub chan PeerConnectedMsg) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}

type PeerDisconnectedMsg struct {
	Id   uuid.UUID
	Time string
}

func WaitForPeerDisconnectionCmd(sub chan PeerDisconnectedMsg) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}

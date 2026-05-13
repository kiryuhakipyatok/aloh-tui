package commands

import tea "github.com/charmbracelet/bubbletea"

type PeerConnectedMsg struct {
	Nickname string
	Time     string
}

func WaitForPeerConnectionCmd(sub chan PeerConnectedMsg) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}

type PeerDisconnectedMsg struct {
	Nickname string
	Time     string
}

func WaitForPeerDisconnectionCmd(sub chan PeerDisconnectedMsg) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}

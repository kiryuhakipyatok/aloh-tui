package commands

import (
	"aloh-tui/internal/sshclient"

	tea "github.com/charmbracelet/bubbletea"
)

func WaitForEventMessageCmd(sub chan sshclient.Event) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}
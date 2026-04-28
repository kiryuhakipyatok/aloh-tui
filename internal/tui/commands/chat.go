package commands

import tea "github.com/charmbracelet/bubbletea"

type ChatMessage struct {
	Time     string
	Nickname string
	Text     string
}

func WaitForChatMessageCmd(sub chan ChatMessage) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}
package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type RawChatMessage struct {
	Time string
	Id   uuid.UUID
	Data []byte
}

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

func WaitForRawChatMessageCmd(sub chan RawChatMessage) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}

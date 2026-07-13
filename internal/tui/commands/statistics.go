package commands

import (
	"aloh-tui/internal/entities/users"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type StatiscticsMsg struct {
	Err error
}

func IncreaseAmountOfMessagesCmd(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := StatiscticsMsg{}
		if err := user.IncreaseAmountOfMessages(); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func IncreaseAmountOfConnectionsCmd(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := StatiscticsMsg{}
		if err := user.IncreaseAmountOfConnections(); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func CountMaxTimeInConnectionCmd(user *users.User, stop chan struct{}) tea.Cmd {
	return func() tea.Msg {
		msg := StatiscticsMsg{}
		if err := user.CountMaxTimeInConnection(stop); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func IncreaseAmountOfConnectionsByUser(user *users.User, id uuid.UUID, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := StatiscticsMsg{}
		if err := user.IncreaseAmountOfConnectionsByUser(id, nickname); err != nil {
			msg.Err = err
		}
		return msg
	}
}

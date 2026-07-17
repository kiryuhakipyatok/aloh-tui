package commands

import (
	"aloh-tui/internal/entities/users"

	tea "github.com/charmbracelet/bubbletea"
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

func IncreaseAmountOfConnectionsByUser(user *users.User, iden users.Identity) tea.Cmd {
	return func() tea.Msg {
		msg := StatiscticsMsg{}
		if err := user.IncreaseAmountOfConnectionsByUser(iden); err != nil {
			msg.Err = err
		}
		return msg
	}
}

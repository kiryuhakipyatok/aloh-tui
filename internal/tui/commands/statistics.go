package commands

import (
	"aloh-tui/internal/entities"

	tea "github.com/charmbracelet/bubbletea"
)

type StatiscticsMsg struct {
	Err error
}

func IncreaseAmountOfFriendsCmd(user *entities.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := StatiscticsMsg{}
		if err := user.IncreaseAmountOfFriends(nickname); err != nil {
			msg.Err = err
		}
		return msg
	}
}

// func DecreaseAmountOfFriendsCmd(user *entities.User) tea.Cmd {
// 	return func() tea.Msg {
// 		msg := StatiscticsMsg{}
// 		if err := user.DecreaseAmountOfFriends(); err != nil {
// 			msg.Err = err
// 		}
// 		return msg
// 	}
// }

func IncreaseAmountOfMessagesCmd(user *entities.User) tea.Cmd {
	return func() tea.Msg {
		msg := StatiscticsMsg{}
		if err := user.IncreaseAmountOfMessages(); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func IncreaseAmountOfConnectionsCmd(user *entities.User) tea.Cmd {
	return func() tea.Msg {
		msg := StatiscticsMsg{}
		if err := user.IncreaseAmountOfConnections(); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func CountMaxTimeInConnectionCmd(user *entities.User, stop chan struct{}) tea.Cmd {
	return func() tea.Msg {
		msg := StatiscticsMsg{}
		if err := user.CountMaxTimeInConnection(stop); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func IncreaseAmountOfConnectionsByUser(user *entities.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := StatiscticsMsg{}
		if err := user.IncreaseAmountOfConnectionsByUser(nickname); err != nil {
			msg.Err = err
		}
		return msg
	}
}

package commands

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/pkg/errs"

	tea "github.com/charmbracelet/bubbletea"
)

type ConnectMsg struct {
	Err error
}

func ConnectToAllUsersCmd(user *users.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := ConnectMsg{}
		if !user.IsFriend(nickname) {
			msg.Err = errs.ErrNotFriend()
			return msg
		}
		if user.Networking != nil {
			if err := user.Networking.ConnectToAllUsers(nickname); err != nil {
				msg.Err = err
			}
		}
		return msg
	}
}

func ConnectToUserCmd(user *users.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := ConnectMsg{}
		if user.Networking != nil {
			if err := user.Networking.ConnectToUser(nickname); err != nil {
				msg.Err = err
			}
		}
		return msg
	}
}

package commands

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/pkg/errs"

	tea "github.com/charmbracelet/bubbletea"
)

type ConnectToUserMsg struct {
	Err error
}

func ConnectToUserCmd(user *users.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := ConnectToUserMsg{}
		if !user.IsFriend(nickname) {
			msg.Err = errs.ErrNotFriend()
			return msg
		}
		if user.Networking != nil {
			if err := user.Networking.ConnectToUser(nickname); err != nil {
				msg.Err = err
			}
		}
		return msg
	}
}

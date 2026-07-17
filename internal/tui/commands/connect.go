package commands

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/pkg/errs"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type ConnectMsg struct {
	Err error
}

func ConnectToAllUsersCmd(user *users.User, iden users.Identity) tea.Cmd {
	return func() tea.Msg {
		msg := ConnectMsg{}
		if !user.IsFriend(iden) {
			msg.Err = errs.ErrNotFriend()
			return msg
		}
		if user.Networking != nil {
			if err := user.Networking.ConnectToAllUsers(iden.ID); err != nil {
				msg.Err = err
			}
		}
		return msg
	}
}

func ConnectToUserCmd(user *users.User, id uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		msg := ConnectMsg{}
		if user.Networking != nil {
			if err := user.Networking.ConnectToUser(id); err != nil {
				msg.Err = err
			}
		}
		return msg
	}
}

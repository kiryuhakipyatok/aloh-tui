package commands

import (
	"aloh-tui/internal/entities"
	"aloh-tui/internal/sshclient"
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type FriendsMsg struct {
	Typee    uint
	Nickname string
	Err      error
}

func SendFriendRequestCmd(user *entities.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee:    sshclient.NEW_FRIEND,
			Nickname: nickname,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := user.SSHClient.NewFriendReq(ctx, nickname); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func AcceptFriendRequestCmd(user *entities.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee:    sshclient.ACCEPT_FRIEND,
			Nickname: nickname,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := user.SSHClient.AcceptFriendReq(ctx, nickname); err != nil {
			msg.Err = err
			return msg
		}
		if err := user.IncreaseAmountOfFriends(nickname); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func DenyFriendRequestCmd(user *entities.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee:    sshclient.DENY_FRIEND,
			Nickname: nickname,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := user.SSHClient.DenyFriendReq(ctx, nickname); err != nil {
			msg.Err = err
		}

		return msg
	}
}

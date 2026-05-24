package commands

import (
	"aloh-tui/internal/entities/users"
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

func SendFriendRequestCmd(user *users.User, nickname string) tea.Cmd {
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

func AcceptFriendRequestCmd(user *users.User, nickname string) tea.Cmd {
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

func DenyFriendRequestCmd(user *users.User, nickname string) tea.Cmd {
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

func DeleteFromFriendsCmd(user *users.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee:    sshclient.DELETE_FRIEND,
			Nickname: nickname,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := user.SSHClient.DeleteFromFriends(ctx, nickname); err != nil {
			msg.Err = err
			return msg
		}
		if err := user.DecreaseAmountOfFriends(nickname); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func BlockUserCmd(user *users.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee:    sshclient.BLOCK_USER,
			Nickname: nickname,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := user.SSHClient.BlockUser(ctx, nickname); err != nil {
			msg.Err = err
			return msg
		}

		if user.IsFriend(nickname) {
			if err := user.DecreaseAmountOfFriends(nickname); err != nil {
				msg.Err = err
			}
		} else {
			user.DeleteFriendReq(nickname)
		}
		return msg
	}
}

func UnblockUserCmd(user *users.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee:    sshclient.UNBLOCK_USER,
			Nickname: nickname,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := user.SSHClient.UnblockUser(ctx, nickname); err != nil {
			msg.Err = err
			return msg
		}

		return msg
	}
}

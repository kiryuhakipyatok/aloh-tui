package commands

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/sshclient"
	"aloh-tui/pkg/errs"
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type FriendsMsg struct {
	Typee    uint
	Nickname string
	Err      error
}

func NewFriendReqCmd(user *users.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee:    sshclient.NEW_FRIEND_REQ,
			Nickname: nickname,
		}
		if user.IsFriend(nickname) {
			msg.Err = errs.ErrAlreadyExists()
			return msg
		}

		user.NewFriendReq(nickname)
		return msg
	}
}

func NewFriendCmd(user *users.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee:    sshclient.ACCEPT_FRIEND,
			Nickname: nickname,
		}
		if user.IsFriend(nickname) {
			msg.Err = errs.ErrAlreadyExists()
			return msg
		}
		user.NewFriend(nickname)
		return msg
	}
}

func SendFriendRequestCmd(user *users.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee:    sshclient.SEND_FRIEND_REQ,
			Nickname: nickname,
		}
		if user.IsBlocked(nickname) {
			msg.Err = errs.ErrNotFound()
			return msg
		}
		if user.IsFriend(nickname) {
			msg.Err = errs.ErrAlreadyExists()
			return msg
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := user.SSHClient.NewFriendReq(ctx, nickname); err != nil {
			msg.Err = err
			return msg
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
		if user.IsFriend(nickname) {
			msg.Err = errs.ErrAlreadyExists()
			return msg
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := user.SSHClient.AcceptFriendReq(ctx, nickname); err != nil {
			msg.Err = err
			return msg
		}
		user.NewFriend(nickname)
		user.DeleteFriendReq(nickname)
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
			return msg
		}
		user.DeleteFriendReq(nickname)
		return msg
	}
}

func DeleteFromFriendsCmd(user *users.User, nickname string, isInitiator bool) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee:    sshclient.DELETE_FRIEND,
			Nickname: nickname,
		}
		if !user.IsFriend(nickname) {
			msg.Err = errs.ErrNotFriend()
			return msg
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if isInitiator {
			if err := user.SSHClient.DeleteFromFriends(ctx, nickname); err != nil {
				msg.Err = err
				return msg
			}
		}
		if err := user.DeleteFriend(nickname); err != nil {
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
		if user.IsBlocked(nickname) {
			msg.Err = errs.ErrAlreadyExists()
			return msg
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := user.SSHClient.BlockUser(ctx, nickname); err != nil {
			msg.Err = err
			return msg
		}
		user.BlockUser(nickname)
		return msg
	}
}

func UnblockUserCmd(user *users.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee:    sshclient.UNBLOCK_USER,
			Nickname: nickname,
		}
		if !user.IsBlocked(nickname) {
			msg.Err = errs.ErrNotBlocked()
			return msg
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := user.SSHClient.UnblockUser(ctx, nickname); err != nil {
			msg.Err = err
			return msg
		}
		user.UnblockUser(nickname)
		return msg
	}
}

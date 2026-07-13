package commands

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/sshclient"
	"aloh-tui/pkg/errs"
	"context"
	"encoding/json"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type FriendsMsg struct {
	Typee    uint
	Id       uuid.UUID
	Nickname string
	Err      error
}

func NewFriendReqCmd(user *users.User, id uuid.UUID, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee:    sshclient.NEW_FRIEND_REQ,
			Id:       id,
			Nickname: nickname,
		}
		if user.IsFriend(id) {
			msg.Err = errs.ErrAlreadyExists()
			return msg
		}

		frReq := users.FriendReq{
			FriendPersonal: users.FriendPersonal{
				ID:       id,
				Nickname: nickname,
			},
		}

		user.NewFriendReq(frReq)
		return msg
	}
}

func NewFriendCmd(user *users.User, id uuid.UUID, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee:    sshclient.ACCEPT_FRIEND,
			Id:       id,
			Nickname: nickname,
		}
		if user.IsFriend(id) {
			msg.Err = errs.ErrAlreadyExists()
			return msg
		}
		fr := users.Friend{
			FriendPersonal: users.FriendPersonal{
				ID:       id,
				Nickname: nickname,
			},
		}
		user.NewFriend(fr)
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
		// if user.IsFriend(id) {
		// 	msg.Err = errs.ErrAlreadyExists()
		// 	return msg
		// }
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := user.SSHClient.NewFriendReq(ctx, []byte(nickname)); err != nil {
			msg.Err = err
			return msg
		}
		return msg
	}
}

func AcceptFriendRequestCmd(user *users.User, id uuid.UUID, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee:    sshclient.ACCEPT_FRIEND,
			Id:       id,
			Nickname: nickname,
		}
		if user.IsFriend(id) {
			msg.Err = errs.ErrAlreadyExists()
			return msg
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := user.SSHClient.AcceptFriendReq(ctx, []byte(nickname)); err != nil {
			msg.Err = err
			return msg
		}
		//user.NewFriend(fr)
		//user.DeleteFriendReq(fr.ID)
		return msg
	}
}

func DenyFriendRequestCmd(user *users.User, id uuid.UUID, nickname string) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee:    sshclient.DENY_FRIEND,
			Id:       id,
			Nickname: nickname,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := user.SSHClient.DenyFriendReq(ctx, []byte(nickname)); err != nil {
			msg.Err = err
			return msg
		}
		user.DeleteFriendReq(id)
		return msg
	}
}

func DeleteFromFriendsCmd(user *users.User, id uuid.UUID, nickname string, isInitiator bool) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee:    sshclient.DELETE_FRIEND,
			Nickname: nickname,
		}
		if !user.IsFriend(id) {
			msg.Err = errs.ErrNotFriend()
			return msg
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if isInitiator {
			if err := user.SSHClient.DeleteFromFriends(ctx, []byte(nickname)); err != nil {
				msg.Err = err
				return msg
			}
		}
		if err := user.DeleteFriend(id, nickname); err != nil {
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
		if err := user.SSHClient.BlockUser(ctx, []byte(nickname)); err != nil {
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
		if err := user.SSHClient.UnblockUser(ctx, []byte(nickname)); err != nil {
			msg.Err = err
			return msg
		}
		user.UnblockUser(nickname)
		return msg
	}
}

func UpdateCurrentConnectsCmd(user *users.User, conns []string) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee: sshclient.FRIEND_CONNECTIONS,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		data, err := json.Marshal(&conns)
		if err != nil {
			msg.Err = err
			return msg
		}
		if err := user.SSHClient.UpdateCurrentConnects(ctx, data); err != nil {
			msg.Err = err
		}
		return msg
	}
}

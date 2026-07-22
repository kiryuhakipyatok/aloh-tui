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
	Identity users.Identity
	Typee    uint
	Err      error
}

func NewFriendReqCmd(user *users.User, iden users.Identity) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Identity: iden,
			Typee:    sshclient.NEW_FRIEND_REQ,
		}
		if user.IsFriend(iden) {
			msg.Err = errs.ErrAlreadyExists()
			return msg
		}

		frReq := users.FriendReq{
			Identity: iden,
		}

		user.NewFriendReq(frReq)
		return msg
	}
}

func NewFriendCmd(user *users.User, iden users.Identity) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Identity: iden,
			Typee:    sshclient.ACCEPT_FRIEND,
		}
		if user.IsFriend(iden) {
			msg.Err = errs.ErrAlreadyExists()
			return msg
		}
		fr := users.Friend{
			Identity: iden,
		}
		user.NewFriend(fr)
		return msg
	}
}

func SendFriendRequestCmd(user *users.User, iden users.Identity) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Identity: iden,
			Typee:    sshclient.SEND_FRIEND_REQ,
		}
		if user.IsBlocked(iden) {
			msg.Err = errs.ErrNotFound()
			return msg
		}
		if user.IsFriend(iden) {
			msg.Err = errs.ErrAlreadyExists()
			return msg
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := user.SSHClient.NewFriendReq(ctx, iden.Nickname); err != nil {
			msg.Err = err
			return msg
		}
		return msg
	}
}

func AcceptFriendRequestCmd(user *users.User, iden users.Identity) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee:    sshclient.ACCEPT_FRIEND,
			Identity: iden,
		}
		if user.IsFriend(iden) {
			msg.Err = errs.ErrAlreadyExists()
			return msg
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := user.SSHClient.AcceptFriendReq(ctx, iden); err != nil {
			msg.Err = err
			return msg
		}
		fr := users.Friend{
			Identity: iden,
		}
		user.NewFriend(fr)
		//	user.DeleteFriendReq(fr.ID)
		return msg
	}
}

func DenyFriendRequestCmd(user *users.User, iden users.Identity) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee:    sshclient.DENY_FRIEND,
			Identity: iden,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := user.SSHClient.DenyFriendReq(ctx, iden.ID); err != nil {
			msg.Err = err
			return msg
		}
		user.DeleteFriendReq(iden.ID)
		return msg
	}
}

func DeleteFromFriendsCmd(user *users.User, iden users.Identity, isInitiator bool) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Identity: iden,
			Typee:    sshclient.DELETE_FRIEND,
		}
		if !user.IsFriend(iden) {
			msg.Err = errs.ErrNotFriend()
			return msg
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if isInitiator {
			if err := user.SSHClient.DeleteFromFriends(ctx, iden.ID); err != nil {
				msg.Err = err
				return msg
			}
		}
		if err := user.DeleteFriend(iden); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func BlockUserCmd(user *users.User, iden users.Identity) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Identity: iden,
			Typee:    sshclient.BLOCK_USER,
		}
		if user.IsBlocked(iden) {
			msg.Err = errs.ErrAlreadyExists()
			return msg
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		idBytes, err := user.SSHClient.BlockUser(ctx, iden.Nickname)
		if err != nil {
			msg.Err = err
			return msg
		}
		var id uuid.UUID
		if err := json.Unmarshal(idBytes, &id); err != nil {
			msg.Err = err
			return msg
		}

		iden.ID = id

		if err := user.BlockUser(iden); err != nil {
			msg.Err = err
		}
		msg.Identity = iden
		return msg
	}
}

func UnblockUserCmd(user *users.User, iden users.Identity) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee:    sshclient.UNBLOCK_USER,
			Identity: iden,
		}
		if !user.IsBlocked(iden) {
			msg.Err = errs.ErrNotBlocked()
			return msg
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		idBytes, err := user.SSHClient.UnblockUser(ctx, iden.Nickname)
		if err != nil {
			msg.Err = err
			return msg
		}
		var id uuid.UUID
		if err := json.Unmarshal(idBytes, &id); err != nil {
			msg.Err = err
			return msg
		}
		iden.ID = id
		msg.Identity = iden
		user.UnblockUser(iden)
		return msg
	}
}

func UpdateCurrentConnectsCmd(user *users.User, conns []users.Identity) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee: sshclient.FRIEND_CONNECTIONS,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		if err := user.SSHClient.UpdateCurrentConnects(ctx, conns); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func UpdateFriendsTaglineCmd(user *users.User, iden users.Identity, newTagline string) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee:    sshclient.UPDATE_TAGLINE,
			Identity: iden,
		}

		if err := user.UpdateFriendTagline(iden.ID, newTagline); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func UpdateFriendsColorCmd(user *users.User, iden users.Identity, newColor string) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee:    sshclient.UPDATE_COLOR,
			Identity: iden,
		}

		if err := user.UpdateFriendColor(iden, newColor); err != nil {
			msg.Err = err
		}

		return msg
	}
}

func UpdateForeignNicknameCmd(user *users.User, iden users.Identity, newNickname string) tea.Cmd {
	return func() tea.Msg {
		msg := FriendsMsg{
			Typee: sshclient.UPDATE_NICKNAME,
		}

		if err := user.UpdateForeignNickname(iden, newNickname); err != nil {
			msg.Err = err
		}
		iden.Nickname = newNickname
		msg.Identity = iden
		return msg
	}
}

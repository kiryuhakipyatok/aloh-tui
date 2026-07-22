package commands

import (
	"aloh-tui/internal/entities/users"
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	NICKNAME = iota
	TAGLINE
	PASSWORD
	COLOR
)

type AccountMsg struct {
	Typee uint
	Err   error
}

func NewNicknameCmd(user *users.User, newNickname string, password []byte) tea.Cmd {
	return func() tea.Msg {
		msg := AccountMsg{
			Typee: NICKNAME,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := user.SSHClient.NewNickname(ctx, newNickname, password); err != nil {
			msg.Err = err
			return msg
		}
		if err := user.NewNickname(newNickname); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func ChangeTaglineCmd(user *users.User, newTagline string) tea.Cmd {
	return func() tea.Msg {
		msg := AccountMsg{
			Typee: TAGLINE,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := user.SSHClient.SetTagline(ctx, newTagline); err != nil {
			msg.Err = err
			return msg
		}
		if err := user.ChangeTagline(newTagline); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func NewPasswordCmd(user *users.User, oldPassword, newPassword []byte) tea.Cmd {
	return func() tea.Msg {
		msg := AccountMsg{
			Typee: PASSWORD,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := user.SSHClient.NewPassword(ctx, oldPassword, newPassword); err != nil {
			msg.Err = err
		}
		return msg
	}
}

func ChangeColorCmd(user *users.User, newColor string) tea.Cmd {
	return func() tea.Msg {
		msg := AccountMsg{
			Typee: COLOR,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := user.SSHClient.SetColor(ctx, newColor); err != nil {
			msg.Err = err
			return msg
		}
		if err := user.ChangeColor(newColor); err != nil {
			msg.Err = err
		}
		return msg
	}
}

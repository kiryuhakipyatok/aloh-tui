package commands

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/sshclient"
	"aloh-tui/internal/tui/helpers"
	"aloh-tui/pkg/logger"

	tea "github.com/charmbracelet/bubbletea"
)

type AuthMsg struct {
	Typee uint
	Err   error
}

func AuthCmd(user *users.User, eventsChan chan sshclient.Event, appLogger *logger.Logger) tea.Cmd {
	return func() tea.Msg {
		msg := AuthMsg{
			Typee: sshclient.DEFAULT,
		}

		as := helpers.AuthSetup{
			User:       user,
			Typee:      sshclient.DEFAULT,
			EventsChan: eventsChan,
			Log:        appLogger,
			Password:   nil,
		}

		if err := helpers.SetupAuth(as); err != nil {
			msg.Err = err
		}

		return msg
	}
}

func RegisterCmd(user *users.User, eventsChan chan sshclient.Event, appLogger *logger.Logger, password []byte) tea.Cmd {
	return func() tea.Msg {
		msg := AuthMsg{
			Typee: sshclient.REGISTER,
		}
		as := helpers.AuthSetup{
			User:       user,
			Typee:      sshclient.REGISTER,
			EventsChan: eventsChan,
			Log:        appLogger,
			Password:   password,
		}

		if err := helpers.SetupAuth(as); err != nil {
			msg.Err = err
		}

		return msg
	}
}

func LoginCmd(user *users.User, eventsChan chan sshclient.Event, appLogger *logger.Logger, password []byte) tea.Cmd {
	return func() tea.Msg {
		msg := AuthMsg{
			Typee: sshclient.LOGIN,
		}
		as := helpers.AuthSetup{
			User:       user,
			Typee:      sshclient.LOGIN,
			EventsChan: eventsChan,
			Log:        appLogger,
			Password:   password,
		}

		if err := helpers.SetupAuth(as); err != nil {
			msg.Err = err
		}
		return msg
	}
}

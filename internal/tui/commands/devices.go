package commands

import (
	"aloh-tui/internal/entities/users"

	tea "github.com/charmbracelet/bubbletea"
)

type ChangeMicrophoneMessage struct {
	Err error
}

func ChangeMicrophoneCmd(user *users.User, microphone string) tea.Cmd {
	return func() tea.Msg {
		msg := ChangeMicrophoneMessage{}
		if err := user.Engines.AudioEngine.ChangeMicrophone(microphone); err != nil {
			msg.Err = err
			return msg
		}

		if err := user.ChangeMicrophone(microphone); err != nil {
			msg.Err = err
		}

		return msg
	}
}

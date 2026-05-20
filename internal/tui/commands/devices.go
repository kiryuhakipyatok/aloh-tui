package commands

import (
	"aloh-tui/internal/entities"

	tea "github.com/charmbracelet/bubbletea"
)

type ChangeMicrophoneMessage struct {
	Err error
}

func ChangeMicrophoneCmd(user *entities.User, microphone string) tea.Cmd {
	return func() tea.Msg {
		msg := ChangeMicrophoneMessage{}
		if err := user.Engines.AudioEngine.ChangeMicrophone(microphone); err != nil {
			msg.Err = err
			return msg
		}

		user.Data.Devices.Microphone = microphone

		if err := user.UpdateUserJSON(); err != nil {
			msg.Err = err
		}

		return msg
	}
}

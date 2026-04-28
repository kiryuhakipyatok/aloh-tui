package commands

import (
	"aloh-tui/internal/entities"
	"encoding/json"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type ChangeMicrophoneMessage struct {
	Err error
}

func ChangeMicrophoneCmd(user *entities.User, microphone string) tea.Cmd {
	return func() tea.Msg {
		if err := user.Engines.AudioEngine.ChangeMicrophone(microphone); err != nil {
			return ChangeMicrophoneMessage{err}
		}

		user.Data.Devices.Microphone = microphone

		userData, err := json.Marshal(user.Data)
		if err != nil {
			return ChangeMicrophoneMessage{err}
		}

		if err := os.WriteFile(user.Paths.DataFilePath, userData, 0644); err != nil {
			return ChangeMicrophoneMessage{err}
		}

		return ChangeMicrophoneMessage{nil}
	}
}

package commands

import (
	"aloh-tui/internal/entities/users"

	tea "github.com/charmbracelet/bubbletea"
)

type ChangeDeviceMessage struct {
	Err error
}

func ChangeMicrophoneCmd(user *users.User, microphone string) tea.Cmd {
	return func() tea.Msg {
		msg := ChangeDeviceMessage{}
		if user.Engines.AudioEngine != nil {
			if err := user.Engines.AudioEngine.ChangeMicrophone(microphone); err != nil {
				msg.Err = err
				return msg
			}

			if err := user.ChangeMicrophone(microphone); err != nil {
				msg.Err = err
			}
		}

		return msg
	}
}

func ChangeHeadphonesCmd(user *users.User, headphone string) tea.Cmd {
	return func() tea.Msg {
		msg := ChangeDeviceMessage{}
		if user.Engines.AudioEngine != nil {
			if err := user.Engines.AudioEngine.ChangeHeadphones(headphone); err != nil {
				msg.Err = err
				return msg
			}

			if err := user.ChangeHeadphones(headphone); err != nil {
				msg.Err = err
			}
		}

		return msg
	}
}
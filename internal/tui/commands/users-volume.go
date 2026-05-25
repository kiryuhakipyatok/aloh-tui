package commands

import (
	"aloh-tui/internal/entities/users"

	tea "github.com/charmbracelet/bubbletea"
)

type UsersVolumeMsg struct {
	Err error
}

func SetUserVolumeCmd(user *users.User, nickname string, volumeCoeficent float32) tea.Cmd {
	return func() tea.Msg {
		msg := UsersVolumeMsg{}
		if user.Engines.AudioEngine != nil {
			user.Engines.AudioEngine.SetVolume(nickname, volumeCoeficent)

			if err := user.SetUsersVolume(nickname, volumeCoeficent); err != nil {
				msg.Err = err
			}

		}

		return msg
	}
}

func SetupUserVolumeCmd(user *users.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		var vc float32 = 1
		if user.Engines.AudioEngine != nil {
			us := user.GetUsersSetup(nickname)
			if us != nil {
				vc = us.VolumeCoefficient
			}
			user.Engines.AudioEngine.SetVolume(nickname, vc)
		}
		return UsersVolumeMsg{}
	}
}

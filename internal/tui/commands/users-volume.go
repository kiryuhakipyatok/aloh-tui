package commands

import (
	"aloh-tui/internal/entities/users"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type UsersVolumeMsg struct {
	Err error
}

func SetUserVolumeCmd(user *users.User, id uuid.UUID, nickname string, volumeCoeficent float32) tea.Cmd {
	return func() tea.Msg {
		msg := UsersVolumeMsg{}
		if user.Engines.AudioEngine != nil {
			user.Engines.AudioEngine.SetVolume(id, volumeCoeficent)

			if err := user.SetUsersVolume(nickname, volumeCoeficent); err != nil {
				msg.Err = err
			}

		}

		return msg
	}
}

func SetupUserVolumeCmd(user *users.User, id uuid.UUID, nickname string) tea.Cmd {
	return func() tea.Msg {
		var vc float32 = 1
		msg := UsersVolumeMsg{}
		if user.Engines.AudioEngine != nil {
			us, err := user.GetUsersSetup(nickname)
			if err != nil {
				msg.Err = err
				return msg
			}

			vc = us.VolumeCoefficient

			user.Engines.AudioEngine.SetVolume(id, vc)
		}
		return msg
	}
}

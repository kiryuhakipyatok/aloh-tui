package commands

import (
	"aloh-tui/internal/entities/users"

	tea "github.com/charmbracelet/bubbletea"
)

type UsersVolumeMsg struct {
	Err error
}

func SetUserVolumeCmd(user *users.User, iden users.Identity, volumeCoeficent float32) tea.Cmd {
	return func() tea.Msg {
		msg := UsersVolumeMsg{}
		if user.Engines.AudioEngine != nil {
			user.Engines.AudioEngine.SetVolume(iden.ID, volumeCoeficent)

			if err := user.SetUsersVolume(iden.Nickname, volumeCoeficent); err != nil {
				msg.Err = err
			}

		}

		return msg
	}
}

func SetupUserVolumeCmd(user *users.User, iden users.Identity) tea.Cmd {
	return func() tea.Msg {
		var vc float32 = 1
		msg := UsersVolumeMsg{}
		if user.Engines.AudioEngine != nil {
			us, err := user.GetUsersSetup(iden.Nickname)
			if err != nil {
				msg.Err = err
				return msg
			}

			vc = us.VolumeCoefficient

			user.Engines.AudioEngine.SetVolume(iden.ID, vc)
		}
		return msg
	}
}
package commands

import (
	"aloh-tui/internal/entities"

	tea "github.com/charmbracelet/bubbletea"
)

type UsersVolumeMsg struct {
	Err error
}

func SetUserVolumeCmd(user *entities.User, nickname string, volumeCoeficent float32) tea.Cmd {
	return func() tea.Msg {
		msg := UsersVolumeMsg{}
		user.Engines.AudioEngine.SetVolume(nickname, volumeCoeficent)

		us, ok := user.Data.Setup.UsersSetup[nickname]
		if !ok {
			us = &entities.UsersSetup{
				VolumeCoefficient: 1,
			}
		}

		us.VolumeCoefficient = volumeCoeficent

		user.Data.Setup.UsersSetup[nickname] = us

		if err := user.UpdateUserJSON(); err != nil {
			msg.Err = err
		}

		return msg
	}
}

func SetupUserVolumeCmd(user *entities.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		var vc float32 = 1
		us, ok := user.Data.Setup.UsersSetup[nickname]
		if ok {
			vc = us.VolumeCoefficient
		}
		user.Engines.AudioEngine.SetVolume(nickname, vc)
		return UsersVolumeMsg{}
	}
}

package commands

import (
	"aloh-tui/internal/entities"
	"encoding/json"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type UsersVolumeMsg struct {
	Err error
}

func SetUserVolumeCmd(user *entities.User, nickname string, volumeCoeficent float32) tea.Cmd {
	return func() tea.Msg {
		user.Engines.AudioEngine.SetVolume(nickname, volumeCoeficent)

		us, ok := user.Data.Setup.UsersSetup[nickname]
		if !ok {
			us = entities.UsersSetup{}
		}

		us.VolumeCoefficient = volumeCoeficent

		user.Data.Setup.UsersSetup[nickname] = us

		userData, err := json.Marshal(user.Data)
		if err != nil {
			return UsersVolumeMsg{Err: err}
		}

		if err := os.WriteFile(user.Paths.DataFilePath, userData, 0644); err != nil {
			return UsersVolumeMsg{Err: err}
		}

		return UsersVolumeMsg{Err: nil}
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

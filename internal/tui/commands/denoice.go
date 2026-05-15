package commands

import (
	"aloh-tui/internal/entities"
	"encoding/json"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type OnOffDenoiceMsg struct {
	Err error
}

func OnOffHardDenoiceCmd(user *entities.User) tea.Cmd {
	return func() tea.Msg {
		h:= user.Engines.AudioEngine.OnOffHardDenoice()
		user.Data.Setup.HardDenoise = h
	
		userData, err := json.Marshal(user.Data)
		if err != nil {
			return OnOffDenoiceMsg{Err: err}
		}

		if err := os.WriteFile(user.Paths.DataFilePath, userData, 0644); err != nil {
			return OnOffDenoiceMsg{Err: err}
		}
		return OnOffDenoiceMsg{
			Err: nil,
		}
	}
}

func OnOffSoftDenoiceCmd(user *entities.User) tea.Cmd {
	return func() tea.Msg {
		s := user.Engines.AudioEngine.OnOffSoftDenoice()
		user.Data.Setup.SoftDenoise = s

		userData, err := json.Marshal(user.Data)
		if err != nil {
			return OnOffDenoiceMsg{Err: err}
		}

		if err := os.WriteFile(user.Paths.DataFilePath, userData, 0644); err != nil {
			return OnOffDenoiceMsg{Err: err}
		}
		return OnOffDenoiceMsg{
			Err: nil,
		}
	}
}

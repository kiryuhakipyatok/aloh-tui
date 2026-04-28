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

func OnOffDenoiceCmd(user *entities.User) tea.Cmd {
	return func() tea.Msg {
		d := user.Engines.AudioEngine.OnOffDenoice()
		user.Data.Setup.Denoise = d

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

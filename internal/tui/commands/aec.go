package commands

import (
	"aloh-tui/internal/entities"
	"encoding/json"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type OnOffAECMsg struct {
	Err error
}

func OnOffAECCmd(user *entities.User) tea.Cmd {
	return func() tea.Msg {
		d := user.Engines.AudioEngine.OnOffAEC()
		user.Data.Setup.AEC = d

		userData, err := json.Marshal(user.Data)
		if err != nil {
			return OnOffAECMsg{Err: err}
		}

		if err := os.WriteFile(user.Paths.DataFilePath, userData, 0644); err != nil {
			return OnOffAECMsg{Err: err}
		}
		return OnOffAECMsg{
			Err: nil,
		}
	}
}

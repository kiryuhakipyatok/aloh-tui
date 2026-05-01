package commands

import (
	"aloh-tui/internal/entities"
	"encoding/json"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type OnOffFilterMsg struct {
	Err error
}

func OnOffFilterCmd(user *entities.User) tea.Cmd {
	return func() tea.Msg {
		d := user.Engines.AudioEngine.OnOffFilter()
		user.Data.Setup.Filter = d

		userData, err := json.Marshal(user.Data)
		if err != nil {
			return OnOffFilterMsg{Err: err}
		}

		if err := os.WriteFile(user.Paths.DataFilePath, userData, 0644); err != nil {
			return OnOffFilterMsg{Err: err}
		}
		return OnOffFilterMsg{
			Err: nil,
		}
	}
}

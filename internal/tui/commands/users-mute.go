package commands

import (
	"aloh-tui/internal/entities"
	"encoding/json"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type MuteUnmuteUserMsg struct {
	Err error
}

func MuteUnmuteUserCmd(user *entities.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		res, err := user.Engines.AudioEngine.MuteUnmuteUser(nickname)
		if err != nil {
			return MuteUnmuteUserMsg{Err: err}
		}

		us, ok := user.Data.Setup.UsersSetup[nickname]
		if !ok {
			us = entities.UsersSetup{}
		}

		us.Muted = res

		user.Data.Setup.UsersSetup[nickname] = us

		userData, err := json.Marshal(user.Data)
		if err != nil {
			return MuteUnmuteUserMsg{Err: err}
		}

		if err := os.WriteFile(user.Paths.DataFilePath, userData, 0644); err != nil {
			return MuteUnmuteUserMsg{Err: err}
		}

		return MuteUnmuteUserMsg{Err: nil}
	}

}

func SetupUserMuteCmd(user *entities.User, nickname string) tea.Cmd {
	return func() tea.Msg {
		var state bool
		us, ok := user.Data.Setup.UsersSetup[nickname]
		if ok {
			state = us.Muted
		}
		user.Engines.AudioEngine.SetMuteState(nickname, state)
	
		return MuteUnmuteUserMsg{}
	}
}

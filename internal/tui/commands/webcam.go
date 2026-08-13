package commands

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/media/video"

	tea "github.com/charmbracelet/bubbletea"
)

type OnOffWindowWebcamMsg struct {
	Err error
}

func OnOffWindowWebcamCmd(ve video.VideoEngine, iden users.Identity) tea.Cmd {
	return func() tea.Msg {
		msg := OnOffWindowWebcamMsg{}
		_, err := ve.OnOffUsersWindow(iden.ID, iden.Nickname)
		if err != nil {
			msg.Err = err
		}
		return msg
	}
}

func OnOffUserWindowWebcamCmd(ve video.VideoEngine) tea.Cmd {
	return func() tea.Msg {
		msg := OnOffWindowWebcamMsg{}
		_, err := ve.OnOffWindow()
		if err != nil {
			msg.Err = err
		}
		return msg
	}
}

package commands

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/media/video"

	tea "github.com/charmbracelet/bubbletea"
)

type OnOffVideoMsg struct {
	Err error
}

func OnOffWindowWebcamCmd(ve video.VideoEngine, iden users.Identity) tea.Cmd {
	return func() tea.Msg {
		msg := OnOffVideoMsg{}
		_, err := ve.OnOffUsersWebcamWindow(iden.ID, iden.Nickname)
		if err != nil {
			msg.Err = err
		}
		return msg
	}
}

func OnOffUserWindowWebcamCmd(ve video.VideoEngine) tea.Cmd {
	return func() tea.Msg {
		msg := OnOffVideoMsg{}
		_, err := ve.OnOffWebcamWindow()
		if err != nil {
			msg.Err = err
		}
		return msg
	}
}

func OnOffWindowScreenCmd(ve video.VideoEngine, iden users.Identity) tea.Cmd {
	return func() tea.Msg {
		msg := OnOffVideoMsg{}
		_, err := ve.OnOffUsersScreenWindow(iden.ID, iden.Nickname)
		if err != nil {
			msg.Err = err
		}
		return msg
	}
}

func OnOffUserWindowScreenCmd(ve video.VideoEngine) tea.Cmd {
	return func() tea.Msg {
		msg := OnOffVideoMsg{}
		_, err := ve.OnOffScreenWindow()
		if err != nil {
			msg.Err = err
		}
		return msg
	}
}

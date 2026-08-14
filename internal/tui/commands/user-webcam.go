package commands

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/networking"

	tea "github.com/charmbracelet/bubbletea"
)

type UserVideo struct {
	Err error
}

func OnOffWebcamCmd(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := UserVideo{}
		ve := user.Engines.VideoEngine
		if ve != nil {
			res, err := ve.OnOffWebcam()
			if err != nil {
				msg.Err = err
				return msg
			}
			netw := user.Networking
			if netw != nil {
				webcamEvent, err := networking.WebcamEvent(res)
				if err != nil {
					msg.Err = err
					return msg
				}
				if err := netw.NewEvent(webcamEvent); err != nil {
					msg.Err = err
				}
			}
		}

		return msg
	}
}

func OnOffScreenCmd(user *users.User) tea.Cmd {
	return func() tea.Msg {
		msg := UserVideo{}
		ve := user.Engines.VideoEngine
		if ve != nil {
			res, err := ve.OnOffScreen()
			if err != nil {
				msg.Err = err
				return msg
			}
			netw := user.Networking
			if netw != nil {
				screenEvent, err := networking.ScreenEvent(res)
				if err != nil {
					msg.Err = err
					return msg
				}
				if err := netw.NewEvent(screenEvent); err != nil {
					msg.Err = err
				}
			}
		}

		return msg
	}
}

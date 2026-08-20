package commands

import (
	"aloh-tui/internal/media/video"

	tea "github.com/charmbracelet/bubbletea"
)

type KeyFrameMsg struct {
	Err error
}

func SendWebcamKeyFrameCmd(ve video.VideoEngine) tea.Cmd {
	return func() tea.Msg {
		msg := KeyFrameMsg{}
		if ve != nil {
			if err := ve.SendWebcamKeyFrame(); err != nil {
				msg.Err = err
			}
		}

		return msg
	}
}

func SendScreenKeyFrameCmd(ve video.VideoEngine) tea.Cmd {
	return func() tea.Msg {
		msg := KeyFrameMsg{}
		if ve != nil {

			if err := ve.SendScreenKeyFrame(); err != nil {
				msg.Err = err
			}

		}

		return msg
	}
}

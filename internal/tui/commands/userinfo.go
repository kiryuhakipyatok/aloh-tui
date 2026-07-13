package commands

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/networking"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type UserInfoMsg struct {
	Err error
}

func SendUserInfo(user *users.User, id uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		msg := UserInfoMsg{}
		if user.Networking != nil {
			audioSetup := user.GetAudio()
			e, err := networking.GeneralEvent(audioSetup.Mutes.FullMute, audioSetup.Mutes.MicMute,
				audioSetup.Denoises.HardDenoise, audioSetup.Denoises.SoftDenoise)
			if err != nil {
				msg.Err = err
				return msg
			}
			if err := user.Networking.NewEvent(e); err != nil {
				msg.Err = err
				return err
			}
		}
		return msg
	}
}

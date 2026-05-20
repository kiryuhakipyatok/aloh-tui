package commands

import (
	"aloh-tui/internal/auth"
	"aloh-tui/internal/entities"
	"aloh-tui/internal/media/audio"
	"aloh-tui/internal/networking"
	"aloh-tui/pkg/errs"
	"aloh-tui/pkg/logger"
	"encoding/json"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
)

type AuthMsg struct {
	Typee uint
	Err   error
}

func AuthCmd(user *entities.User, appLogger *logger.Logger, password []byte) tea.Cmd {
	return func() tea.Msg {
		msg := AuthMsg{
			Typee: auth.DEFAULT,
		}

		payload, err := auth.Auth(user.Data.Personal.Nickname, user.Paths.KeysPath, auth.DEFAULT, password)
		if err != nil {
			msg.Err = err
			return msg
		}

		pd := entities.Personal{}
		if err := json.Unmarshal(payload, &pd); err != nil {
			msg.Err = err
			return msg
		}

		user.Data.Personal = pd

		netwroking, err := networking.NewNetworking(user.Data.Personal.Nickname, user.Paths.LogFilePath)
		if err != nil {
			msg.Err = err
			return msg
		}

		audioEngine, err := audio.NewAudioEngine(appLogger, audio.AudioSetup{
			Microphone:  user.Data.Devices.Microphone,
			Aec:         user.Data.Setup.AEC,
			HardDenoice: user.Data.Setup.HardDenoise,
			SoftDenoice: user.Data.Setup.SoftDenoise,
			Filtered:    user.Data.Setup.Filter,
		})
		if err != nil {
			msg.Err = err
			return msg
		}

		if err := audioEngine.SetNetworking(netwroking); err != nil {
			msg.Err = err
			return msg
		}

		user.Engines.AudioEngine = audioEngine
		user.Networking = netwroking

		if err := user.UpdateUserJSON(); err != nil {
			msg.Err = err
			return msg
		}

		return msg
	}
}

func RegisterCmd(user *entities.User, appLogger *logger.Logger, password, repPassword []byte) tea.Cmd {
	return func() tea.Msg {
		msg := AuthMsg{
			Typee: auth.REGISTER,
		}
		if !slices.Equal(password, repPassword) {
			msg.Err = errs.ErrPasswordsNotEqual
			return msg
		}
		if _, err := auth.Auth(user.Data.Personal.Nickname, user.Paths.KeysPath, auth.REGISTER, password); err != nil {
			msg.Err = err
			return msg
		}
		user.Data.Setup.SoftDenoise = true
		netwroking, err := networking.NewNetworking(user.Data.Personal.Nickname, user.Paths.LogFilePath)
		if err != nil {
			msg.Err = err
			return msg
		}

		audioEngine, err := audio.NewAudioEngine(appLogger, audio.AudioSetup{
			Microphone:  user.Data.Devices.Microphone,
			Aec:         user.Data.Setup.AEC,
			HardDenoice: user.Data.Setup.HardDenoise,
			SoftDenoice: true,
			Filtered:    user.Data.Setup.Filter,
		})
		if err != nil {
			msg.Err = err
			return msg
		}

		user.Data.Devices.Microphone = audioEngine.GetCurrentMicrophone().Name

		if err := audioEngine.SetNetworking(netwroking); err != nil {
			msg.Err = err
			return msg
		}

		user.Engines.AudioEngine = audioEngine
		user.Networking = netwroking

		if err := user.UpdateUserJSON(); err != nil {
			msg.Err = err
			return msg
		}

		return msg
	}
}

func LoginCmd(user *entities.User, appLogger *logger.Logger, secret []byte) tea.Cmd {
	return func() tea.Msg {
		msg := AuthMsg{
			Typee: auth.LOGIN,
		}
		regTime, err := auth.Auth(user.Data.Personal.Nickname, user.Paths.KeysPath, auth.LOGIN, secret)
		if err != nil {
			msg.Err = err
			return msg
		}
		user.Data.Personal.RegisterTime = string(regTime)
		user.Data.Setup.SoftDenoise = true

		netwroking, err := networking.NewNetworking(user.Data.Personal.Nickname, user.Paths.LogFilePath)
		if err != nil {
			msg.Err = err
			return msg
		}

		audioEngine, err := audio.NewAudioEngine(appLogger, audio.AudioSetup{
			Microphone:  user.Data.Devices.Microphone,
			Aec:         user.Data.Setup.AEC,
			HardDenoice: user.Data.Setup.HardDenoise,
			SoftDenoice: true,
			Filtered:    user.Data.Setup.Filter,
		})
		if err != nil {
			msg.Err = err
			return msg
		}
		user.Data.Devices.Microphone = audioEngine.GetCurrentMicrophone().Name
		if err := audioEngine.SetNetworking(netwroking); err != nil {
			msg.Err = err
			return msg
		}

		user.Engines.AudioEngine = audioEngine
		user.Networking = netwroking

		if err := user.UpdateUserJSON(); err != nil {
			msg.Err = err
		}

		return msg
	}
}

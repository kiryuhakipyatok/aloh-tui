package commands

import (
	"aloh-tui/internal/auth"
	"aloh-tui/internal/entities"
	"aloh-tui/internal/media/audio"
	"aloh-tui/internal/networking"
	"aloh-tui/pkg/errs"
	"aloh-tui/pkg/logger"
	"encoding/json"
	"os"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
)

type AuthMsg struct {
	Typee uint
	Err   error
}

func AuthCmd(user *entities.User, appLogger *logger.Logger, password []byte) tea.Cmd {
	return func() tea.Msg {
		payload, err := auth.Auth(user.Data.Personal.Nickname, user.Paths.KeysPath, auth.DEFAULT, password)
		if err != nil {
			return AuthMsg{Typee: auth.DEFAULT, Err: err}
		}

		pd := entities.Personal{}
		if err := json.Unmarshal(payload, &pd); err != nil {
			return AuthMsg{Typee: auth.DEFAULT, Err: err}
		}

		user.Data.Personal = pd

		netwroking, err := networking.NewNetworking(user.Data.Personal.Nickname, user.Paths.LogFilePath)
		if err != nil {
			return AuthMsg{Typee: auth.DEFAULT, Err: err}
		}

		audioEngine, err := audio.NewAudioEngine(appLogger, audio.AudioSetup{
			Microphone:  user.Data.Devices.Microphone,
			Aec:         user.Data.Setup.AEC,
			HardDenoice: user.Data.Setup.HardDenoise,
			SoftDenoice: user.Data.Setup.SoftDenoise,
			Filtered:    user.Data.Setup.Filter,
		})
		if err != nil {
			return AuthMsg{Typee: auth.DEFAULT, Err: err}
		}

		if err := audioEngine.SetNetworking(netwroking); err != nil {
			return AuthMsg{Typee: auth.DEFAULT, Err: err}
		}

		user.Engines.AudioEngine = audioEngine
		user.Networking = netwroking

		userData, err := json.Marshal(user.Data)
		if err != nil {
			return AuthMsg{Typee: auth.DEFAULT, Err: err}
		}

		if err := os.WriteFile(user.Paths.DataFilePath, userData, 0644); err != nil {
			return AuthMsg{Typee: auth.DEFAULT, Err: err}
		}

		return AuthMsg{Typee: auth.DEFAULT, Err: nil}
	}
}

func RegisterCmd(user *entities.User, appLogger *logger.Logger, password, repPassword []byte) tea.Cmd {
	return func() tea.Msg {
		if !slices.Equal(password, repPassword) {
			return AuthMsg{Typee: auth.REGISTER, Err: errs.ErrPasswordsNotEqual}
		}
		if _, err := auth.Auth(user.Data.Personal.Nickname, user.Paths.KeysPath, auth.REGISTER, password); err != nil {
			return AuthMsg{Typee: auth.REGISTER, Err: err}
		}
		user.Data.Setup.SoftDenoise = true
		netwroking, err := networking.NewNetworking(user.Data.Personal.Nickname, user.Paths.LogFilePath)
		if err != nil {
			return AuthMsg{Typee: auth.REGISTER, Err: err}
		}

		audioEngine, err := audio.NewAudioEngine(appLogger, audio.AudioSetup{
			Microphone:  user.Data.Devices.Microphone,
			Aec:         user.Data.Setup.AEC,
			HardDenoice: user.Data.Setup.HardDenoise,
			SoftDenoice: true,
			Filtered:    user.Data.Setup.Filter,
		})
		if err != nil {
			return AuthMsg{Typee: auth.REGISTER, Err: err}
		}

		user.Data.Devices.Microphone = audioEngine.GetCurrentMicrophone().Name

		if err := audioEngine.SetNetworking(netwroking); err != nil {
			return AuthMsg{Typee: auth.REGISTER, Err: err}
		}

		user.Engines.AudioEngine = audioEngine
		user.Networking = netwroking

		userData, err := json.Marshal(user.Data)
		if err != nil {
			return AuthMsg{Typee: auth.REGISTER, Err: err}
		}

		if err := os.WriteFile(user.Paths.DataFilePath, userData, 0644); err != nil {
			return AuthMsg{Typee: auth.REGISTER, Err: err}
		}

		return AuthMsg{Typee: auth.REGISTER, Err: nil}
	}
}

func LoginCmd(user *entities.User, appLogger *logger.Logger, secret []byte) tea.Cmd {
	return func() tea.Msg {
		regTime, err := auth.Auth(user.Data.Personal.Nickname, user.Paths.KeysPath, auth.LOGIN, secret)
		if err != nil {
			return AuthMsg{Typee: auth.LOGIN, Err: err}
		}
		user.Data.Personal.RegisterTime = string(regTime)
		user.Data.Setup.SoftDenoise = true

		netwroking, err := networking.NewNetworking(user.Data.Personal.Nickname, user.Paths.LogFilePath)
		if err != nil {
			return AuthMsg{Typee: auth.LOGIN, Err: err}
		}

		audioEngine, err := audio.NewAudioEngine(appLogger, audio.AudioSetup{
			Microphone:  user.Data.Devices.Microphone,
			Aec:         user.Data.Setup.AEC,
			HardDenoice: user.Data.Setup.HardDenoise,
			SoftDenoice: true,
			Filtered:    user.Data.Setup.Filter,
		})
		if err != nil {
			return AuthMsg{Typee: auth.LOGIN, Err: err}
		}
		user.Data.Devices.Microphone = audioEngine.GetCurrentMicrophone().Name
		if err := audioEngine.SetNetworking(netwroking); err != nil {
			return AuthMsg{Typee: auth.LOGIN, Err: err}
		}

		user.Engines.AudioEngine = audioEngine
		user.Networking = netwroking

		userData, err := json.Marshal(user.Data)
		if err != nil {
			return AuthMsg{Typee: auth.LOGIN, Err: err}
		}

		if err := os.WriteFile(user.Paths.DataFilePath, userData, 0644); err != nil {
			return AuthMsg{Typee: auth.LOGIN, Err: err}
		}

		return AuthMsg{Typee: auth.LOGIN, Err: nil}
	}
}

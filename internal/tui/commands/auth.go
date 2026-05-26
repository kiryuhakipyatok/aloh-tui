package commands

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/media/audio"
	"aloh-tui/internal/networking"
	"aloh-tui/internal/sshclient"
	"aloh-tui/pkg/errs"
	"aloh-tui/pkg/logger"
	"context"
	"encoding/json"
	"slices"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type AuthMsg struct {
	Typee uint
	Err   error
}

func AuthCmd(user *users.User, eventsChan chan sshclient.Event, appLogger *logger.Logger, password []byte) tea.Cmd {
	return func() tea.Msg {
		msg := AuthMsg{
			Typee: sshclient.DEFAULT,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		client, personalData, err := sshclient.AuthSSHClient(ctx, appLogger, sshclient.SSHClientSetup{
			Nickname:   user.Data.Personal.Nickname,
			KeysPath:   user.Paths.KeysPath,
			Typee:      sshclient.DEFAULT,
			EventsChan: eventsChan,
			Password:   password,
		})
		if err != nil {
			msg.Err = err
			return msg
		}
		var pd struct {
			Nickname     string               `json:"nickname"`
			RegisterTime time.Time            `json:"registerTime"`
			FriendsReqs  []users.FriendReq `json:"friendsReqs"`
			Friends      []string             `json:"friends"`
		}

		if err := json.Unmarshal(personalData, &pd); err != nil {
			msg.Err = err
			return msg
		}

		user.Data.Personal.Nickname = pd.Nickname
		user.Data.Personal.RegisterTime = pd.RegisterTime.Local().Format("2006-01-02")
		user.Data.Personal.FriendsReqs = pd.FriendsReqs
		user.Data.Personal.Friends = pd.Friends
		user.SSHClient = client

		netwroking, err := networking.NewNetworking(user.Data.Personal.Nickname, user.Paths.LogFilePath)
		if err != nil {
			msg.Err = err
			return msg
		}

		audioEngine, err := audio.NewAudioEngine(appLogger, audio.AudioSetup{
			Microphone:  user.Data.Devices.Microphone,
			Aec:         user.Data.Setup.Audio.AEC,
			HardDenoice: user.Data.Setup.Audio.HardDenoise,
			SoftDenoice: user.Data.Setup.Audio.SoftDenoise,
			Filtered:    user.Data.Setup.Audio.Filter,
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

func RegisterCmd(user *users.User, eventsChan chan sshclient.Event, appLogger *logger.Logger, password, repPassword []byte) tea.Cmd {
	return func() tea.Msg {
		msg := AuthMsg{
			Typee: sshclient.REGISTER,
		}
		if !slices.Equal(password, repPassword) {
			msg.Err = errs.ErrPasswordsNotEqual
			return msg
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		client, _, err := sshclient.AuthSSHClient(ctx, appLogger, sshclient.SSHClientSetup{
			Nickname:   user.Data.Personal.Nickname,
			KeysPath:   user.Paths.KeysPath,
			Typee:      sshclient.REGISTER,
			EventsChan: eventsChan,
			Password:   password,
		})
		if err != nil {
			msg.Err = err
			return msg
		}
		user.Data.Setup.Audio.SoftDenoise = true
		user.SSHClient = client
		netwroking, err := networking.NewNetworking(user.Data.Personal.Nickname, user.Paths.LogFilePath)
		if err != nil {
			msg.Err = err
			return msg
		}

		audioEngine, err := audio.NewAudioEngine(appLogger, audio.AudioSetup{
			Microphone:  user.Data.Devices.Microphone,
			Aec:         user.Data.Setup.Audio.AEC,
			HardDenoice: user.Data.Setup.Audio.HardDenoise,
			SoftDenoice: true,
			Filtered:    user.Data.Setup.Audio.Filter,
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

func LoginCmd(user *users.User, eventsChan chan sshclient.Event, appLogger *logger.Logger, secret []byte) tea.Cmd {
	return func() tea.Msg {
		msg := AuthMsg{
			Typee: sshclient.LOGIN,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		client, personalData, err := sshclient.AuthSSHClient(ctx, appLogger, sshclient.SSHClientSetup{
			Nickname:   user.Data.Personal.Nickname,
			KeysPath:   user.Paths.KeysPath,
			Typee:      sshclient.LOGIN,
			EventsChan: eventsChan,
			Password:   nil,
		})
		if err != nil {
			msg.Err = err
			return msg
		}

		var pd struct {
			Nickname     string               `json:"nickname"`
			RegisterTime time.Time            `json:"registerTime"`
			FriendsReqs  []users.FriendReq `json:"friendsReqs"`
			Friends      []string             `json:"friends"`
		}

		if err := json.Unmarshal(personalData, &pd); err != nil {
			msg.Err = err
			return msg
		}

		user.Data.Personal.Nickname = pd.Nickname
		user.Data.Personal.RegisterTime = pd.RegisterTime.Local().Format("2006-01-02")
		user.Data.Personal.FriendsReqs = pd.FriendsReqs
		user.Data.Personal.Friends = pd.Friends
		user.SSHClient = client

		user.Data.Setup.Audio.SoftDenoise = true

		netwroking, err := networking.NewNetworking(user.Data.Personal.Nickname, user.Paths.LogFilePath)
		if err != nil {
			msg.Err = err
			return msg
		}

		audioEngine, err := audio.NewAudioEngine(appLogger, audio.AudioSetup{
			Microphone:  user.Data.Devices.Microphone,
			Aec:         user.Data.Setup.Audio.AEC,
			HardDenoice: user.Data.Setup.Audio.HardDenoise,
			SoftDenoice: true,
			Filtered:    user.Data.Setup.Audio.Filter,
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

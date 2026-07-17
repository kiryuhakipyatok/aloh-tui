package helpers

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/media/audio"
	"aloh-tui/internal/networking"
	"aloh-tui/internal/sshclient"
	"aloh-tui/pkg/errs"
	"aloh-tui/pkg/logger"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type AuthSetup struct {
	Typee      uint
	User       *users.User
	EventsChan chan sshclient.Event
	Password   []byte
	Log        *logger.Logger
}

func SetupAuth(as AuthSetup) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	client, personalData, err := sshclient.AuthSSHClient(ctx, as.Log, sshclient.SSHClientSetup{
		Nickname:   as.User.Data.Identity.Nickname,
		KeysPath:   as.User.Paths.KeysPath,
		Typee:      as.Typee,
		EventsChan: as.EventsChan,
		Password:   as.Password,
	})
	if err != nil {
		if !errors.Is(err, errs.ErrAuth()) {
			return err
		}
	} else {
		var pd struct {
			Identity     users.Identity    `json:"identity"`
			Tagline      string            `json:"tagline"`
			RegisterTime time.Time         `json:"registerTime"`
			FriendsReqs  []users.FriendReq `json:"friendsReqs"`
			Friends      []users.Friend    `json:"friends"`
			BlockedUsers []users.Identity  `json:"blocked-users"`
		}

		if err := json.Unmarshal(personalData, &pd); err != nil {
			return err
		}

		fr := make(map[uuid.UUID]users.Friend)
		for _, f := range pd.Friends {
			fr[f.ID] = f
		}

		as.User.Data.Identity = pd.Identity
		as.User.Data.Personal.RegisterTime = pd.RegisterTime.Format("2006-01-02")
		as.User.Data.Personal.FriendsReqs = pd.FriendsReqs
		as.User.Data.Personal.Friends = fr
		as.User.Data.Personal.BlockedUsers = pd.BlockedUsers
		as.User.SSHClient = client

		networking, err := networking.NewNetworking(as.User.Data.Identity.ID, as.User.Paths.LogFilePath)
		if err != nil {
			return err
		}
		audioEngine, err := audio.NewAudioEngine(as.Log, audio.AudioSetup{
			Microphone:  as.User.Data.Devices.Microphone,
			Aec:         as.User.Data.Setup.Audio.AEC,
			HardDenoice: as.User.Data.Setup.Audio.Denoises.HardDenoise,
			SoftDenoice: as.User.Data.Setup.Audio.Denoises.SoftDenoise,
			Filtered:    as.User.Data.Setup.Audio.Filter,
		})
		if err != nil {
			return err
		}

		if err := audioEngine.SetNetworking(networking); err != nil {
			return err
		}

		as.User.SSHClient = client
		as.User.Engines.AudioEngine = audioEngine
		as.User.Networking = networking

		if err := as.User.UpdateUserJSON(); err != nil {
			return err
		}

		as.Log.Info("users ps", pd)
	}
	return nil
}

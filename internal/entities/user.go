package entities

import (
	"aloh-tui/internal/media/audio"
	"aloh-tui/internal/networking"
	"aloh-tui/internal/sshclient"
	"encoding/json"
	"os"
	"sync"
	"time"
)

type User struct {
	Data       Data
	Paths      Paths
	Networking networking.Networking
	Engines    Engines
	SSHClient  sshclient.SSHClient
	mu         sync.Mutex
}

func NewUser(logFilePath, keysPath, dataFilePath, defColor string) *User {
	return &User{
		Paths: Paths{
			LogFilePath:  logFilePath,
			KeysPath:     keysPath,
			DataFilePath: dataFilePath,
		},
		Data: Data{
			Setup: Setup{
				UsersSetup:           make(map[string]*UsersSetup, 5),
				ThemeColor:           defColor,
				AudioNotifications:   true,
				DesktopNotifications: true,
				BestFriendTag:        "👑",
			},
			Statistics: Statistics{
				BestFriend: BestFriend{
					Nickname: "nobody",
				},
			},
		},
	}
}

func (u *User) UpdateUserJSON() error {
	u.mu.Lock()
	defer u.mu.Unlock()
	userData, err := json.Marshal(u.Data)
	if err != nil {
		return err
	}

	if err := os.WriteFile(u.Paths.DataFilePath, userData, 0644); err != nil {
		return err
	}

	return nil
}

func (u *User) IncreaseAmountOfFriends(nickanme string) error {
	u.Data.Personal.Friends = append(u.Data.Personal.Friends, nickanme)
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) IncreaseAmountOfMessages() error {
	u.Data.Statistics.AmountOfMessages++
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) IncreaseAmountOfConnections() error {
	u.Data.Statistics.AmountOfConnections++
	if err := u.UpdateUserJSON(); err != nil {
		return err
	}
	return nil
}

func (u *User) IncreaseAmountOfConnectionsByUser(nickname string) error {
	if uc, ok := u.Data.Setup.UsersSetup[nickname]; ok {
		uc.AmountOfConnections++
		if uc.AmountOfConnections > u.Data.Statistics.BestFriend.AmountOfConnections {
			u.Data.Statistics.BestFriend.AmountOfConnections = uc.AmountOfConnections
			u.Data.Statistics.BestFriend.Nickname = nickname
		}
		if err := u.UpdateUserJSON(); err != nil {
			return err
		}
	}
	return nil
}

func (u *User) CountMaxTimeInConnection(stop chan struct{}) error {
	var currentTime uint
	defer func() error {
		if currentTime > u.Data.Statistics.MaxTimeInConnetion {
			u.Data.Statistics.MaxTimeInConnetion = currentTime
			if err := u.UpdateUserJSON(); err != nil {
				return err
			}
		}
		return nil
	}()
	for {
		select {
		case <-stop:
			return nil
		case <-time.After(time.Minute * 1):
			currentTime++
			u.Data.Statistics.AmountOfMinutesInConnections++
			if err := u.UpdateUserJSON(); err != nil {
				return err
			}
		}
	}
}

type Data struct {
	Personal   Personal   `json:"personalData"`
	Devices    Devices    `json:"devices"`
	Setup      Setup      `json:"setup"`
	Statistics Statistics `json:"statistics"`
}

type Personal struct {
	Nickname     string `json:"nickname"`
	RegisterTime string `json:"registerTime"`
	FriendsReqs  []FriendReq
	Friends      []string
}

type FriendReq struct {
	Nickname string    `json:"nickname"`
	ReqTime  time.Time `json:"reqTime"`
}

type Paths struct {
	LogFilePath  string
	KeysPath     string
	DataFilePath string
}

type Engines struct {
	AudioEngine audio.AudioEngine
}

type Devices struct {
	Microphone string `json:"microphone"`
	Headphones string `json:"headphones"`
}

type Setup struct {
	HardDenoise          bool                   `json:"hard-denoise"`
	SoftDenoise          bool                   `json:"soft-denoise"`
	AEC                  bool                   `json:"aec"`
	Filter               bool                   `json:"filter"`
	UsersSetup           map[string]*UsersSetup `json:"users-setup"`
	ThemeColor           string                 `json:"theme-color"`
	DesktopNotifications bool                   `json:"desktop-notifications"`
	AudioNotifications   bool                   `json:"audio-notifications"`
	BestFriendTag        string                 `json:"best-friend-tag"`
}

type UsersSetup struct {
	VolumeCoefficient   float32 `json:"volume-coeficent"`
	Muted               bool    `json:"muted"`
	AmountOfConnections uint    `json:"amount-of-connections"`
}

type Statistics struct {
	AmountOfFriends              uint       `json:"amount-of-friends"`
	MaxTimeInConnetion           uint       `json:"max-time-in-connections"`
	AmountOfMessages             uint       `json:"amount-of-messages"`
	AmountOfMinutesInConnections uint       `json:"minutes-in-connections"`
	AmountOfConnections          uint       `json:"amount-of-connections"`
	BestFriend                   BestFriend `json:"best-friend"`
}

type BestFriend struct {
	Nickname            string `json:"nickname"`
	AmountOfConnections uint   `json:"amount-of-connections"`
}

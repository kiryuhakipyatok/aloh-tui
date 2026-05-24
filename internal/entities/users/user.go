package users

import (
	"aloh-tui/internal/media/audio"
	"aloh-tui/internal/networking"
	"aloh-tui/internal/sshclient"
	"sync"
	"time"
)

type User struct {
	Data       Data
	Paths      Paths
	Networking networking.Networking
	Engines    Engines
	SSHClient  sshclient.SSHClient
	mu         sync.RWMutex
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
				AppNotifications:     true,
			},
			Statistics: Statistics{
				BestFriend: BestFriend{
					Nickname: "nobody",
				},
			},
			Personal: Personal{
				Friends:     make([]string, 0, 5),
				FriendsReqs: make([]FriendReq, 0, 5),
			},
		},
	}
}

type Data struct {
	Personal   Personal   `json:"personalData"`
	Devices    Devices    `json:"devices"`
	Setup      Setup      `json:"setup"`
	Statistics Statistics `json:"statistics"`
}

type Personal struct {
	Nickname     string      `json:"nickname"`
	RegisterTime string      `json:"registerTime"`
	FriendsReqs  []FriendReq `json:"-"`
	Friends      []string    `json:"-"`
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
	AppNotifications     bool                   `json:"app-notifications"`
	BestFriendTag        string                 `json:"best-friend-tag"`
	NotificaionSign      string                 `json:"notification-sign"`
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

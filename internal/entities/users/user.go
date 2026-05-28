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

func NewUser(logFilePath, keysPath, dataFilePath string) *User {
	return &User{
		Paths: Paths{
			LogFilePath:  logFilePath,
			KeysPath:     keysPath,
			DataFilePath: dataFilePath,
		},
		Data: Data{
			Setup: Setup{
				Audio: Audio{
					UsersSetup:  make(map[string]*UsersSetup, 5),
					SoftDenoise: true,
				},
				Notifications: Notifications{
					AudioNotifications:   true,
					DesktopNotifications: true,
					AppNotifications:     true,
				},
				Appereance: Appereance{
					ShowTime: true,
					ShowDate: true,
					ShowZone: true,
				},
			},
			Statistics: Statistics{
				BestFriend: noBF(),
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
	BlockedUsers []string    `json:"-"`
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
	Audio         Audio         `json:"audio"`
	Appereance    Appereance    `json:"appereance"`
	Notifications Notifications `json:"notifications"`
}

type Appereance struct {
	BestFriendTag  string `json:"best-friend-tag"`
	NotificaionTag string `json:"notification-tag"`
	ThemeColor     string `json:"theme-color"`
	ShowTime       bool   `json:"show-time"`
	ShowDate       bool   `json:"show-date"`
	ShowZone       bool   `json:"show-zone"`
}

type Notifications struct {
	DesktopNotifications bool `json:"desktop-notifications"`
	AudioNotifications   bool `json:"audio-notifications"`
	AppNotifications     bool `json:"app-notifications"`
}

type Audio struct {
	HardDenoise bool                   `json:"hard-denoise"`
	SoftDenoise bool                   `json:"soft-denoise"`
	AEC         bool                   `json:"aec"`
	Filter      bool                   `json:"filter"`
	UsersSetup  map[string]*UsersSetup `json:"users-setup"`
}

type UsersSetup struct {
	VolumeCoefficient   float32 `json:"volume-coeficent"`
	Muted               bool    `json:"muted"`
	AmountOfConnections uint    `json:"amount-of-connections"`
}

type Statistics struct {
	MinutesInCurrentConnection   uint       `json:"minutes-in-current-connection"`
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

func noBF() BestFriend {
	return BestFriend{
		Nickname: "nobody",
	}
}

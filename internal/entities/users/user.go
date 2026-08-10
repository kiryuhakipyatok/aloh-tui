package users

import (
	"aloh-tui/internal/entities/setups"
	"aloh-tui/internal/media/audio"
	"aloh-tui/internal/media/video"
	"aloh-tui/internal/networking"
	"aloh-tui/internal/sshclient"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	DEF_TM     = "#A6E22E"
	DEF_BFTAG  = "👑"
	DEF_NOTTAG = "🔔"
	DEF_BANTAG = "🚫"
	DEF_COLOR  = "#random"
)

type User struct {
	Data       Data                  `json:"data"`
	Paths      Paths                 `json:"paths"`
	Networking networking.Networking `json:"-"`
	Engines    Engines               `json:"-"`
	SSHClient  sshclient.SSHClient   `json:"-"`
	mu         sync.RWMutex          `json:"-"`
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
					Denoises: Denoises{
						SoftDenoise: true,
					},
					UsersSetup: make(map[string]setups.UsersSetup, 5),
				},
				Notifications: Notifications{
					AudioNotifications:   true,
					DesktopNotifications: true,
					AppNotifications:     true,
				},
				Binds: Binds{
					FriendsTab:  "ALT+F",
					ChatTab:     "ALT+C",
					VoiceTab:    "ALT+G",
					VideoTab:    "ALT+D",
					ProfileTab:  "ALT+E",
					SettingsTab: "ALT+S",
					MicMute:     "ALT+V",
					FullMute:    "ALT+B",
					UserMute:    "ALT+Z",
				},
				Appereance: Appereance{
					BestFriendTag:  DEF_BFTAG,
					ThemeColor:     DEF_TM,
					NotificaionTag: DEF_NOTTAG,
					BanTag:         DEF_BANTAG,
				},
			},
			Account: Account{
				Color: DEF_COLOR,
			},
			Statistics: Statistics{
				BestFriend: noBF(),
			},
			Personal: Personal{
				Friends:     make(map[uuid.UUID]Friend, 5),
				FriendsReqs: make([]FriendReq, 0, 5),
			},
		},
	}
}

type Data struct {
	Identity   Identity   `json:"identity"`
	Personal   Personal   `json:"personal"`
	Devices    Devices    `json:"devices"`
	Setup      Setup      `json:"setup"`
	Account    Account    `json:"account"`
	Statistics Statistics `json:"statistics"`
}

type Identity = sshclient.Identity

type Personal struct {
	RegisterTime string               `json:"registerTime"`
	FriendsReqs  []FriendReq          `json:"-"`
	Friends      map[uuid.UUID]Friend `json:"-"`
	BlockedUsers []Identity           `json:"-"`
}

type FriendReq struct {
	Identity Identity  `json:"identity"`
	ReqTime  time.Time `json:"reqTime"`
}

type Paths struct {
	LogFilePath  string
	KeysPath     string
	DataFilePath string
}

type Engines struct {
	AudioEngine audio.AudioEngine
	VideoEngine video.VideoEngine
}

type Devices struct {
	Microphone string `json:"microphone"`
	Headphones string `json:"headphones"`
}

type Setup struct {
	Audio         Audio         `json:"audio"`
	Appereance    Appereance    `json:"appereance"`
	Notifications Notifications `json:"notifications"`
	Binds         Binds         `json:"binds"`
}

type Binds struct {
	FriendsTab  string `json:"friends-tab"`
	ChatTab     string `json:"chat-tab"`
	VoiceTab    string `json:"voice-tab"`
	VideoTab    string `json:"video-tab"`
	ProfileTab  string `json:"profile-tab"`
	SettingsTab string `json:"settings-tab"`
	MicMute     string `json:"mic-mute"`
	FullMute    string `json:"full-mute"`
	UserMute    string `json:"user-mute"`
}

type Appereance struct {
	BestFriendTag  string `json:"best-friend-tag"`
	NotificaionTag string `json:"notification-tag"`
	BanTag         string `json:"ban-tag"`
	ThemeColor     string `json:"theme-color"`
	ShowTime       bool   `json:"show-time"`
	ShowDate       bool   `json:"show-date"`
	ShowZone       bool   `json:"show-zone"`
}

type Account struct {
	Tagline string `json:"tagline"`
	Color   string `json:"color"`
}

type Notifications struct {
	DesktopNotifications bool `json:"desktop-notifications"`
	AudioNotifications   bool `json:"audio-notifications"`
	AppNotifications     bool `json:"app-notifications"`
}

type Audio struct {
	AEC      bool     `json:"aec"`
	Denoises Denoises `json:"denoises"`
	Mutes    Mutes    `json:"mutes"`
	Filter   bool     `json:"filter"`

	UsersSetup map[string]setups.UsersSetup `json:"users-setup"`
}

type Denoises struct {
	HardDenoise bool `json:"hard-denoise"`
	SoftDenoise bool `json:"soft-denoise"`
}

type Mutes struct {
	FullMute bool `json:"full-mute"`
	MicMute  bool `json:"mic-mute"`
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
	Identity            Identity `json:"identity"`
	AmountOfConnections uint     `json:"amount-of-connections"`
}

func noBF() BestFriend {
	return BestFriend{
		Identity: Identity{
			Nickname: "nobody",
		},
	}
}

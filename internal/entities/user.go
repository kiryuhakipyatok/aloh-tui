package entities

import (
	"aloh-tui/internal/media/audio"
	"aloh-tui/internal/networking"
)

type User struct {
	Data       Data
	Paths      Paths
	Networking networking.Networking
	Engines    Engines
}

type Data struct {
	Personal   Personal `json:"personalData"`
	Devices    Devices  `json:"devices"`
	Setup      Setup    `json:"setup"`
	Statistics Statistics
}

type Personal struct {
	Nickname     string `json:"nickname"`
	RegisterTime string `json:"registerTime"`
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
	HardDenoise          bool                  `json:"hard-denoise"`
	SoftDenoise          bool                  `json:"soft-denoise"`
	AEC                  bool                  `json:"aec"`
	Filter               bool                  `json:"filter"`
	UsersSetup           map[string]UsersSetup `json:"users-setup"`
	ThemeColor           string                `json:"theme-color"`
	DesktopNotifications bool                  `json:"desktop-notifications"`
	AudioNotifications   bool                  `json:"audio-notifications"`
}

type UsersSetup struct {
	VolumeCoefficient float32 `json:"volume-coeficent"`
	Muted             bool    `json:"muted"`
}

type Statistics struct {
	AmountOfFriends              uint         `json:"amount-of-friends"`
	MaxTimeInConnetion           uint         `json:"max-time-in-connections"`
	AmountOfMessages             uint         `json:"amount-of-messages"`
	AmountOfMinutesInConnections uint         `json:"minutes-in-connections"`
	FavoriteUser                 FavoriteUser `json:"favorite-user"`
	FavoriteMsg                  FavoriteMsg  `json:"favorite-msg"`
}

type FavoriteMsg struct {
	Msg             string `json:"msg"`
	AmountOfSending uint   `json:"amount-of-sending"`
}

type FavoriteUser struct {
	Nickname            string `json:"nickname"`
	AmountOfConnections uint   `json:"amount-of-connections"`
}
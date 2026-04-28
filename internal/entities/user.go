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
	Personal Personal `json:"personalData"`
	Devices  Devices  `json:"devices"`
	Setup    Setup    `json:"setup"`
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
	Denoise    bool                  `json:"denoise"`
	AEC        bool                  `json:"aec"`
	UsersSetup map[string]UsersSetup `json:"users-setup"`
}

type UsersSetup struct {
	VolumeCoefficient float32 `json:"volume-coeficent"`
	Muted             bool    `json:"muted"`
}

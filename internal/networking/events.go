package networking

import (
	alohnetwork "github.com/kiryuhakipyatok/aloh-networking"
)

const (
	FULL_MUTE    = alohnetwork.FULL_MUTE
	MIC_MUTE     = alohnetwork.MIC_MUTE
	HARD_DENOISE = alohnetwork.HARD_DENOISE
	SOFT_DENOISE = alohnetwork.SOFT_DENOISE
	GENERAL      = alohnetwork.GENERAL
)

func MuteMicEvent(state bool) (alohnetwork.Event, error) {
	return alohnetwork.MuteMicEvent(state)
}

func MuteFullEvent(state bool) (alohnetwork.Event, error) {
	return alohnetwork.MuteFullEvent(state)
}

func HardDenoiseEvent(state bool) (alohnetwork.Event, error) {
	return alohnetwork.HardDenoiseEvent(state)
}

func SoftDenoiseEvent(state bool) (alohnetwork.Event, error) {
	return alohnetwork.SoftDenoiseEvent(state)
}

func GeneralEvent(fm, mm, hd, sd bool) (alohnetwork.Event, error) {
	return alohnetwork.GeneralEvent(fm, mm, hd, sd)
}

func DataToState(data []byte) (bool, error) {
	return alohnetwork.DataToState(data)
}

func DataToGeneral(data []byte) (alohnetwork.GeneralData, error) {
	return alohnetwork.DataToGeneral(data)
}

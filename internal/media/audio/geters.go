package audio

import "aloh-tui/internal/media"

type Geters interface {
	GetCurrentMicrophone() media.Device
	GetCurrentHeadphones() media.Device
}

func (ae *audioEngine) GetCurrentMicrophone() media.Device {
	ae.mu.RLock()
	curMic := ae.CurrentMicrophone
	ae.mu.RUnlock()
	return curMic
}

func (ae *audioEngine) GetCurrentHeadphones() media.Device {
	ae.mu.RLock()
	curH := ae.CurrentHeadphones
	ae.mu.RUnlock()
	return curH
}

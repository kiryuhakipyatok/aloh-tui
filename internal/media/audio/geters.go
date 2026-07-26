package audio

type Geters interface {
	GetCurrentMicrophone() DeviceInfo
	GetCurrentHeadphones() DeviceInfo
}

func (ae *audioEngine) GetCurrentMicrophone() DeviceInfo {
	ae.mu.RLock()
	curMic := ae.CurrentMicrophone
	ae.mu.RUnlock()
	return curMic
}

func (ae *audioEngine) GetCurrentHeadphones() DeviceInfo {
	ae.mu.RLock()
	curH := ae.CurrentHeadphones
	ae.mu.RUnlock()
	return curH
}

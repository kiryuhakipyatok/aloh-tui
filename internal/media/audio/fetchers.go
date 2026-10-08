package audio

import (
	"aloh-tui/internal/media"

	"github.com/google/uuid"
)

type Fetchers interface {
	FetchMicrophones() map[string]media.Device
	FetchHeadphones() map[string]media.Device
	FetchUsersMutes() map[uuid.UUID]struct{}
	FetchSpeakingUsers() map[uuid.UUID]float64
}

func (ae *audioEngine) FetchSpeakingUsers() map[uuid.UUID]float64 {
	ae.mu.RLock()
	defer ae.mu.RUnlock()
	speakers := make(map[uuid.UUID]float64, len(ae.usersAudio))
	for n, ua := range ae.usersAudio {
		if ua.isSpeaking.Load() {
			speakers[n] = ua.rms
		}
	}
	return speakers
}

func (ae *audioEngine) FetchUsersMutes() map[uuid.UUID]struct{} {
	ae.mu.RLock()
	defer ae.mu.RUnlock()
	muters := make(map[uuid.UUID]struct{}, len(ae.usersAudio))
	for n, ua := range ae.usersAudio {
		if ua.muted.Load() {
			muters[n] = struct{}{}
		}
	}
	return muters
}

func (ae *audioEngine) FetchMicrophones() map[string]media.Device {
	ae.mu.RLock()
	mics := ae.Microphones
	ae.mu.RUnlock()
	return mics
}

func (ae *audioEngine) FetchHeadphones() map[string]media.Device {
	ae.mu.RLock()
	heads := ae.Headphones
	ae.mu.RUnlock()
	return heads
}

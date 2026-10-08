package audio

import (
	"aloh-tui/internal/media/audio/casters"
	"aloh-tui/pkg/logger"

	"github.com/google/uuid"
)

type Players interface {
	PlayNotification()
	PlayUserVoice(id uuid.UUID, userVoiceByte []byte)
}

func (ae *audioEngine) PlayNotification() {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	ae.notificationBytes = ae.notificationSound
	ae.notificationPos = 0
}

func (ae *audioEngine) PlayUserVoice(id uuid.UUID, userVoiceByte []byte) {
	if ae.muted.Load() || ae.switching.Load() {
		return
	}
	var err error
	ae.mu.RLock()
	ua, ok := ae.usersAudio[id]
	ae.mu.RUnlock()
	if !ok {
		ua, err = ae.newUserAudio()
		if err != nil {
			return
		}
		ae.mu.Lock()
		ae.usersAudio[id] = ua
		ae.mu.Unlock()
	}
	if ua.muted.Load() {
		return
	}
	ua.mu.Lock()
	v := ua.volumeCoefficient
	pcmBuffer := ae.int16BuffersPool.Get().([]int16)
	n, err := ua.decoder.Decode(userVoiceByte, pcmBuffer)
	if err != nil {
		ae.log.Error(ae.errLogCount, "failed to decode incoming opus packet", logger.Err(err))
		ua.mu.Unlock()
		return
	}

	ua.mu.Unlock()
	setupVolume(v, pcmBuffer[:n])
	ua.mu.Lock()
	casters.Int16ToBytes(pcmBuffer[:n], ua.decodedBuffer[:n*2])

	ua.data = append(ua.data, ua.decodedBuffer[:n*2]...)

	ua.mu.Unlock()
	ae.int16BuffersPool.Put(pcmBuffer[:4096])
	ua.mu.Lock()
	if len(ua.data) > 96000 {
		ua.data = ua.data[len(ua.data)-jitterSize:]
	}
	ua.mu.Unlock()
}

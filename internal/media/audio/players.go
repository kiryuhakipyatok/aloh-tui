package audio

import (
	"aloh-tui/internal/media/audio/casters"
	"aloh-tui/pkg/logger"

	"github.com/google/uuid"
	"gopkg.in/hraban/opus.v2"
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
	ae.mu.Lock()
	ua, ok := ae.usersAudio[id]
	if !ok {
		opusDecoder, err := opus.NewDecoder(48000, 1)
		if err != nil {
			ae.log.Error(ae.errLogCount, "failed to create new opus decoder", logger.Err(err))
			ae.mu.Unlock()
			return
		}

		ua = &usersAudio{
			data:              make([]byte, 0, 48000),
			float32Buffer:     make([]float32, frameLen),
			denoicedBuffer:    make([]float32, frameLen),
			decoder:           opusDecoder,
			volumeCoefficient: 1,
			decodedBuffer:     make([]byte, 5760),
			samples:           make([]int16, frameLen),
		}

		ae.usersAudio[id] = ua
	}
	if ua.muted.Load() {
		ae.mu.Unlock()
		return
	}
	v := ua.volumeCoefficient
	ae.mu.Unlock()
	pcmBuffer := ae.int16BuffersPool.Get().([]int16)
	n, err := ua.decoder.Decode(userVoiceByte, pcmBuffer)
	if err != nil {
		ae.log.Error(ae.errLogCount, "failed to decode incoming opus packet", logger.Err(err))
		return
	}
	setupVolume(v, pcmBuffer[:n])
	ae.mu.Lock()
	casters.Int16ToBytes(pcmBuffer[:n], ua.decodedBuffer[:n*2])

	ua.data = append(ua.data, ua.decodedBuffer[:n*2]...)
	ae.int16BuffersPool.Put(pcmBuffer[:4096])

	if len(ua.data) > 96000 {
		ua.data = ua.data[len(ua.data)-jitterSize:]
	}
	ae.mu.Unlock()
}

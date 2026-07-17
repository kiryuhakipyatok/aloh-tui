package audio

import (
	"aloh-tui/internal/networking"
	"aloh-tui/pkg/logger"
	"errors"

	"github.com/google/uuid"
	"gopkg.in/hraban/opus.v2"
)

type Seters interface {
	SetNetworking(netw networking.Networking) error
	SetConnected()
	SetDisconnected() error
	SetVolume(id uuid.UUID, vc float32)
	SetMuteState(id uuid.UUID, mute bool)
}

func (ae *audioEngine) SetNetworking(netw networking.Networking) error {
	if netw == nil {
		return errors.New("networking is nil")
	}
	ae.mu.Lock()
	ae.netw = netw
	ae.mu.Unlock()
	return nil
}

func (ae *audioEngine) SetConnected() {
	ae.connected.Store(true)
}

func (ae *audioEngine) SetDisconnected() error {
	ae.connected.Store(false)
	ae.voiceHolder.Store(0)
	ae.userIsSpeaking.Store(false)
	// ae.micNativeBuffer = ae.micNativeBuffer[:0]
	// ae.workMic = ae.workMic[:0]
	// ae.resampledWorkMic = ae.resampledWorkMic[:0]
	// ae.monoCaptureBuffer = ae.monoCaptureBuffer
	// ae.mu.Lock()
	// defer ae.mu.Unlock()
	if ae.opusEncoder != nil {
		if err := ae.opusEncoder.Reset(); err != nil {
			return err
		}
	}

	ae.mu.Lock()
	for _, ua := range ae.usersAudio {
		ua.data = ua.data[:0]
		ua.playing = false
		ua.framesCount = 0
		ua.isSpeaking.Store(false)

		if ua.personalPreprocessor != nil {
			if err := ua.personalPreprocessor.Close(); err != nil {
				return err
			}
		}
		if ua.personalHardDenoise != nil {
			if err := ua.personalHardDenoise.Close(); err != nil {
				return err
			}
		}
	}

	clear(ae.usersAudio)

	ae.mu.Unlock()
	return nil
}

func (ae *audioEngine) SetMuteState(id uuid.UUID, mute bool) {
	ae.mu.Lock()
	ua, ok := ae.usersAudio[id]
	if ok {
		ua.muted.Store(mute)
		ae.mu.Unlock()
		return
	}
	ae.mu.Unlock()

	opusDecoder, err := opus.NewDecoder(48000, 1)
	if err != nil {
		ae.log.Error(ae.errLogCount, "failed to create new opus decoder", logger.Err(err))
		return
	}
	newUa := &usersAudio{
		data:              make([]byte, 0, 48000),
		float32Buffer:     make([]float32, frameLen),
		denoicedBuffer:    make([]float32, frameLen),
		decoder:           opusDecoder,
		volumeCoefficient: 1,
		decodedBuffer:     make([]byte, 5760),
		samples:           make([]int16, frameLen),
	}

	ae.mu.Lock()
	defer ae.mu.Unlock()
	if existingUa, ok := ae.usersAudio[id]; ok {
		existingUa.muted.Store(mute)
	} else {
		ae.usersAudio[id] = newUa
	}
}

func (ae *audioEngine) SetVolume(id uuid.UUID, vc float32) {
	ae.mu.Lock()
	ua, ok := ae.usersAudio[id]
	if ok {
		ua.volumeCoefficient = vc
		ae.mu.Unlock()
		return
	}
	ae.mu.Unlock()

	opusDecoder, err := opus.NewDecoder(48000, 1)
	if err != nil {
		ae.log.Error(ae.errLogCount, "failed to create new opus decoder", logger.Err(err))
		return
	}
	newUa := &usersAudio{
		data:              make([]byte, 0, 48000),
		float32Buffer:     make([]float32, frameLen),
		denoicedBuffer:    make([]float32, frameLen),
		decoder:           opusDecoder,
		volumeCoefficient: vc,
		decodedBuffer:     make([]byte, 5760),
		samples:           make([]int16, frameLen),
	}

	ae.mu.Lock()
	defer ae.mu.Unlock()
	if existingUa, ok := ae.usersAudio[id]; ok {
		existingUa.volumeCoefficient = vc
	} else {
		ae.usersAudio[id] = newUa
	}
}

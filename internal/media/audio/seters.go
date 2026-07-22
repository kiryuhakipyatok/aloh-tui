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

	//SetUsersAudio(usersSetups map[uuid.UUID]setups.UsersSetup) error

	SetVolume(id uuid.UUID, vc float32)
	SetMuteState(id uuid.UUID, mute bool)
}

// func (ae *audioEngine) SetUsersAudio(usersSetups map[uuid.UUID]setups.UsersSetup) error {
// 	opusDecoder, err := opus.NewDecoder(sampleRate, 1)
// 	if err != nil {
// 		ae.log.Error(ae.errLogCount, "failed to create new opus decoder", logger.Err(err))
// 		return err
// 	}
// 	newUa := &usersAudio{
// 		data:              make([]byte, 0, sampleRate),
// 		float32Buffer:     make([]float32, frameLen),
// 		denoicedBuffer:    make([]float32, frameLen),
// 		decoder:           opusDecoder,
// 		volumeCoefficient: 1,
// 		decodedBuffer:     make([]byte, 5760),
// 		samples:           make([]int16, frameLen),
// 	}
// 	ae.mu.Lock()
// 	defer ae.mu.Unlock()
// 	for id, us := range usersSetups {
// 		newUa.muted.Store(us.Muted)
// 		newUa.volumeCoefficient = us.VolumeCoefficient

// 		userHardDenoiseState := us.HardDenoise
// 		if userHardDenoiseState {
// 			newUa.personalHardDenoise = rnnoise.NewRNNoise()
// 		} else if !userHardDenoiseState && newUa.personalHardDenoise != nil {
// 			if err := newUa.personalHardDenoise.Close(); err != nil {
// 				return err
// 			}
// 			newUa.personalHardDenoise = nil
// 		}
// 		newUa.hardDenoised.Store(userHardDenoiseState)

// 		userSoftDenoiseState := us.SoftDenoise
// 		if userSoftDenoiseState && newUa.personalPreprocessor == nil {
// 			newUa.personalPreprocessor = speexdsp.NewPreprocessor(sampleRate, frameLen)
// 			newUa.personalPreprocessor.EnableDenoise(true)
// 			newUa.personalPreprocessor.SetEchoCanceller(nil)
// 		} else if !userSoftDenoiseState && newUa.personalPreprocessor != nil {
// 			if err := newUa.personalPreprocessor.Close(); err != nil {
// 				return err
// 			}
// 			newUa.personalPreprocessor = nil
// 		}
// 		newUa.softDenoised.Store(userSoftDenoiseState)

// 		ae.usersAudio[id] = newUa

// 	}
// 	return nil
// }

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
		ae.log.Info(0, "using existing ua in set mute state", ua)
		ua.muted.Store(mute)
		ae.mu.Unlock()
		return
	}
	ae.mu.Unlock()

	ae.mu.Lock()
	defer ae.mu.Unlock()
	if existingUa, ok := ae.usersAudio[id]; ok {
		existingUa.muted.Store(mute)
	} else {
		if _, err := ae.newUserAudio(id); err != nil {
			ae.log.Error(0, "failed to create user audio", logger.Err(err))
			return
		}
	}
	ae.log.Info(0, "usersAudo after set mute state", ae.usersAudio)
}

func (ae *audioEngine) SetVolume(id uuid.UUID, vc float32) {
	ae.mu.Lock()
	ua, ok := ae.usersAudio[id]
	if ok {
		ae.log.Info(0, "using existing ua in set volume", ua)
		ua.volumeCoefficient = vc
		ae.mu.Unlock()
		return
	}
	ae.mu.Unlock()
	if existingUa, ok := ae.usersAudio[id]; ok {
		existingUa.volumeCoefficient = vc
	} else {
		if _, err := ae.newUserAudio(id); err != nil {
			ae.log.Error(0, "failed to create user audio", logger.Err(err))
			return
		}
	}
	ae.log.Info(0, "usersAudo after set volume", ae.usersAudio)
}

func (ae *audioEngine) newUserAudio(id uuid.UUID) (*usersAudio, error) {
	opusDecoder, err := opus.NewDecoder(sampleRate, 1)
	if err != nil {
		ae.log.Error(ae.errLogCount, "failed to create new opus decoder", logger.Err(err))
		return nil, err
	}
	newUa := &usersAudio{
		data:              make([]byte, 0, sampleRate),
		float32Buffer:     make([]float32, frameLen),
		denoicedBuffer:    make([]float32, frameLen),
		decoder:           opusDecoder,
		volumeCoefficient: 1,
		decodedBuffer:     make([]byte, 5760),
		samples:           make([]int16, frameLen),
	}
	if _, ok := ae.usersAudio[id]; !ok {
		ae.usersAudio[id] = newUa
	}
	return newUa, nil
}

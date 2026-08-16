package audio

import (
	"aloh-tui/internal/networking"
	"aloh-tui/pkg/logger"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
	"gopkg.in/hraban/opus.v2"
)

type Seters interface {
	SetNetworking(netw networking.Networking)
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

func (ae *audioEngine) SetNetworking(netw networking.Networking) {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	ae.netw = netw
}

func (ae *audioEngine) SetConnected() {
	ae.connected.Store(true)
}

func (ae *audioEngine) SetDisconnected() error {
	ae.sessionId.Add(1)
	ae.connected.Store(false)
	ae.voiceHolder.Store(0)
	ae.userIsSpeaking.Store(false)
	ae.playbackReady.Store(false)
	ae.captureReady.Store(false)
	ae.aecDiff.Store(0)
	ae.mu.Lock()
	for len(ae.micDataChan) > 0 {
		unusedVoice := <-ae.micDataChan
		ae.bytesBuffersPool.Put(unusedVoice.data[:1000])
	}
	ae.micNativeBuffer = ae.micNativeBuffer[:0]
	clear(ae.workMic)
	ae.resampledWorkMic = ae.resampledWorkMic[:0]
	ae.monoCaptureBuffer = ae.monoCaptureBuffer[:0]

	clear(ae.voiceBuffer)

	ae.playbackNativeBuffer = ae.playbackNativeBuffer[:0]
	ae.resampledWorkMix = ae.resampledWorkMix[:0]

	clear(ae.workMix)
	clear(ae.pcmBuffer)

	ae.notificationBytes = nil
	ae.notificationPos = 0
	ae.mu.Unlock()
	if ae.opusEncoder != nil {
		if err := ae.opusEncoder.Reset(); err != nil {
			return err
		}
	}
	var eg errgroup.Group
	ae.mu.Lock()
	for _, ua := range ae.usersAudio {
		eg.Go(func() error {
			ua.data = ua.data[:0]
			ua.playing = false
			ua.framesCount = 0
			ua.isSpeaking.Store(false)

			if ua.personalPreprocessor != nil {
				if err := ua.personalPreprocessor.Close(); err != nil {
					return err
				}
			}
			ua.personalPreprocessor = nil
			if ua.personalHardDenoise != nil {
				if err := ua.personalHardDenoise.Close(); err != nil {
					return err
				}
			}
			ua.personalHardDenoise = nil
			return nil
		})
	}

	ae.mu.Unlock()
	if err := eg.Wait(); err != nil {
		clear(ae.usersAudio)
		return err
	}
	ae.mu.Lock()
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
		var err error
		ua, err = ae.newUserAudio()
		if err != nil {
			return
		}

		ae.usersAudio[id] = ua
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
	if existingUa, ok := ae.usersAudio[id]; ok {
		existingUa.volumeCoefficient = vc
	} else {
		var err error
		ua, err = ae.newUserAudio()
		if err != nil {
			return
		}
		ae.usersAudio[id] = ua
	}
	ae.mu.Unlock()
	ae.log.Info(0, "usersAudo after set volume", ae.usersAudio)
}

func (ae *audioEngine) newUserAudio() (*usersAudio, error) {
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
	return newUa, nil
}

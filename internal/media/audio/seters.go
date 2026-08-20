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

	SetVolume(id uuid.UUID, vc float32)
	SetMuteState(id uuid.UUID, mute bool)
}

func (ae *audioEngine) SetNetworking(netw networking.Networking) {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	ae.netw = netw
}

func (ae *audioEngine) SetConnected() {
	ae.log.Info(0, "connected")
	ae.connected.Store(true)
}

func (ae *audioEngine) SetDisconnected() error {

	ae.sessionId.Add(1)
	ae.connected.Store(false)
	ae.log.Info(0, "setted disconencted")
	ae.voiceHolder.Store(0)
	ae.userIsSpeaking.Store(false)
	ae.playbackReady.Store(false)
	ae.captureReady.Store(false)
	ae.aecDiff.Store(0)
	var eg errgroup.Group

	ae.mu.RLock()
	usersAudio := ae.usersAudio
	ae.mu.RUnlock()
	for _, ua := range usersAudio {
		eg.Go(func() error {
			ua.mu.Lock()
			defer ua.mu.Unlock()
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

	if err := eg.Wait(); err != nil {
		ae.mu.Lock()
		clear(ae.usersAudio)
		ae.mu.Unlock()
		return err
	}
	ae.mu.Lock()
	for len(ae.micDataChan) > 0 {
		unusedVoice := <-ae.micDataChan
		ae.bytesBuffersPool.Put(unusedVoice.data[:1000])
	}

	ae.micNativeBuffer = ae.micNativeBuffer[:0]
	clear(ae.workMic)
	clear(ae.resampledWorkMic)
	clear(ae.denoicedBuffer)
	clear(ae.float32Buffer)
	clear(ae.echolessBufferInt16s)
	ae.monoCaptureBuffer = ae.monoCaptureBuffer[:0]

	clear(ae.voiceBuffer)

	ae.playbackNativeBuffer = ae.playbackNativeBuffer[:0]
	clear(ae.resampledWorkMix)

	clear(ae.workMix)
	clear(ae.pcmBuffer)

	ae.notificationBytes = nil
	ae.notificationPos = 0
	if ae.opusEncoder != nil {
		if err := ae.opusEncoder.Reset(); err != nil {
			clear(ae.usersAudio)
			ae.mu.Unlock()
			return err
		}
	}
	
	clear(ae.usersAudio)
	ae.mu.Unlock()
	return nil
}

func (ae *audioEngine) SetMuteState(id uuid.UUID, mute bool) {
	ae.mu.RLock()
	ua, ok := ae.usersAudio[id]
	ae.mu.RUnlock()
	if ok {
		ua.muted.Store(mute)
		return
	}

	ae.mu.RLock()
	existingUa, ok := ae.usersAudio[id]
	ae.mu.RUnlock()
	if ok {
		existingUa.muted.Store(mute)
	} else {
		var err error
		ua, err = ae.newUserAudio()
		if err != nil {
			return
		}
		ae.mu.Lock()
		ae.usersAudio[id] = ua
		ae.mu.Unlock()
	}
	ae.log.Info(0, "usersAudo after set mute state", ae.usersAudio)
}

func (ae *audioEngine) SetVolume(id uuid.UUID, vc float32) {
	ae.mu.RLock()
	ua, ok := ae.usersAudio[id]
	ae.mu.RUnlock()
	if ok {
		ae.log.Info(0, "using existing ua in set volume", ua)
		ua.mu.Lock()
		ua.volumeCoefficient = vc
		ua.mu.Unlock()
		return
	}

	ae.mu.RLock()
	existingUa, ok := ae.usersAudio[id]
	ae.mu.RUnlock()
	if ok {
		ua.mu.Lock()
		existingUa.volumeCoefficient = vc
		ua.mu.Unlock()
	} else {
		var err error
		ua, err = ae.newUserAudio()
		if err != nil {
			return
		}
		ae.mu.Lock()
		ae.usersAudio[id] = ua
		ae.mu.Unlock()
	}
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
		workMix:           make([]int16, 4096),
		decoder:           opusDecoder,
		volumeCoefficient: 1,
		decodedBuffer:     make([]byte, 5760),
		samples:           make([]int16, frameLen),
	}
	return newUa, nil
}

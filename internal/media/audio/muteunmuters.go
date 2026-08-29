package audio

import (
	"aloh-tui/pkg/errs"

	"github.com/google/uuid"
)

type MuteUnmuters interface {
	MuteUnmuteMicro() bool
	MuteUnmute() bool

	MuteUnmuteUser(id uuid.UUID) (bool, error)
}

func (ae *audioEngine) MuteUnmuteUser(id uuid.UUID) (bool, error) {
	ae.mu.RLock()
	ua, ok := ae.usersAudio[id]
	ae.mu.RUnlock()
	if !ok {
		return false, errs.ErrNotFound()
	}
	s := ua.muted.Load()
	ua.muted.Store(!s)
//	ua.muteFade.Store(0)
	//	ua.data = ua.data[:0]
	// clear(ua.float32Buffer)
	// clear(ua.denoicedBuffer)
	// clear(ua.workMix)

	// clear(ua.decodedBuffer)
	// clear(ua.samples)
	return !s, nil
}

func (ae *audioEngine) MuteUnmuteMicro() bool {
	s := ae.mutedMicro.Load()
	ae.muted.Store(false)
	ae.mutedMicro.Store(!s)
	return !s
}

func (ae *audioEngine) MuteUnmute() bool {
	s := ae.muted.Load()
	ae.mutedMicro.Store(false)
	ae.muted.Store(!s)
//	ae.muteFade.Store(0)
	ae.mu.Lock()

	// ae.micNativeBuffer = ae.micNativeBuffer[:0]
	// clear(ae.workMic)
	// clear(ae.resampledWorkMic)
	// clear(ae.denoicedBuffer)
	// clear(ae.float32Buffer)
	// clear(ae.echolessBufferInt16s)
	// ae.monoCaptureBuffer = ae.monoCaptureBuffer[:0]

	// clear(ae.voiceBuffer)

	// ae.playbackNativeBuffer = ae.playbackNativeBuffer[:0]
	// clear(ae.resampledWorkMix)

	// clear(ae.workMix)
	// clear(ae.pcmBuffer)

	for _, ua := range ae.usersAudio {
		// ua.data = ua.data[:0]
		// clear(ua.float32Buffer)
		// clear(ua.denoicedBuffer)
		// clear(ua.workMix)

		// clear(ua.decodedBuffer)
		// clear(ua.samples)

		ua.isSpeaking.Store(false)

		//	ua.playing = false
	}
	ae.mu.Unlock()

	return !s
}

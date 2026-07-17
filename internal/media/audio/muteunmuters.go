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
	ae.mu.Lock()
	defer ae.mu.Unlock()
	ua, ok := ae.usersAudio[id]
	if !ok {
		return false, errs.ErrNotFound()
	}
	s := ua.muted.Load()
	ua.muted.Store(!s)
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
	return !s
}

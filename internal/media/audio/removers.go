package audio

import (
	"github.com/google/uuid"
)

type Removers interface {
	RemoveFromUsersAudio(id uuid.UUID) error
}

func (ae *audioEngine) RemoveFromUsersAudio(id uuid.UUID) error {
	ae.mu.RLock()
	ua, ok := ae.usersAudio[id]
	if !ok {
		return nil
	}
	ae.mu.RUnlock()

	ua.mu.Lock()
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
	ua.mu.Unlock()
	ae.mu.Lock()
	delete(ae.usersAudio, id)
	ae.mu.Unlock()
	return nil
}

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
	ae.mu.RUnlock()
	if !ok {
		return nil
	}


	ua.mu.Lock()
	if ua.personalPreprocessor != nil {
		if err := ua.personalPreprocessor.Close(); err != nil {
			ua.mu.Unlock()
			return err
		}
	}
	if ua.personalHardDenoise != nil {
		if err := ua.personalHardDenoise.Close(); err != nil {
			ua.mu.Unlock()
			return err
		}
	}
	ua.mu.Unlock()
	ae.mu.Lock()
	delete(ae.usersAudio, id)
	ae.mu.Unlock()
	return nil
}

package audio

import (
	"aloh-tui/pkg/errs"

	"github.com/google/uuid"
)

type Removers interface {
	RemoveFromUsersAudio(id uuid.UUID) error
}

func (ae *audioEngine) RemoveFromUsersAudio(id uuid.UUID) error {
	ua, ok := ae.usersAudio[id]
	if !ok {
		return errs.ErrNotFound()
	}
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
	delete(ae.usersAudio, id)
	return nil
}

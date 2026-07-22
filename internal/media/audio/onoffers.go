package audio

import (
	"github.com/google/uuid"
	"github.com/kechako/go-speexdsp"
	rnnoise "github.com/kiryuhakipyatok/rnnoise/cmd"
)

type OnOffers interface {
	OnOffHardDenoice() bool
	OnOffSoftDenoice() bool
	OnOffAEC() bool
	OnOffFilter() bool
	OnOffUsersHardDenoise(id uuid.UUID, state bool) error
	OnOffUsersSoftDenoise(id uuid.UUID, state bool) error
}

func (ae *audioEngine) OnOffUsersHardDenoise(id uuid.UUID, state bool) error {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	var err error
	ua, ok := ae.usersAudio[id]
	if !ok {
		ua, err = ae.newUserAudio(id)
		if err != nil {
			return err
		}
	}
	if state && ua.personalHardDenoise == nil {
		ua.personalHardDenoise = rnnoise.NewRNNoise()
	} else if !state && ua.personalHardDenoise != nil {
		if err := ua.personalHardDenoise.Close(); err != nil {
			return err
		}
		ua.personalHardDenoise = nil
	}
	ua.hardDenoised.Store(state)
	return nil
}

func (ae *audioEngine) OnOffUsersSoftDenoise(id uuid.UUID, state bool) error {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	var err error
	ua, ok := ae.usersAudio[id]
	if !ok {
		ua, err = ae.newUserAudio(id)
		if err != nil {
			return err
		}
	}
	//s := ua.softDenoised.Load()
	if state && ua.personalPreprocessor == nil {
		ua.personalPreprocessor = speexdsp.NewPreprocessor(sampleRate, frameLen)
		ua.personalPreprocessor.EnableDenoise(true)
		ua.personalPreprocessor.SetEchoCanceller(nil)
	} else if !state && ua.personalPreprocessor != nil {
		if err := ua.personalPreprocessor.Close(); err != nil {
			return err
		}
		ua.personalPreprocessor = nil
	}
	ua.softDenoised.Store(state)

	return nil
}

func (ae *audioEngine) OnOffHardDenoice() bool {
	h := ae.hardDenoiced.Load()
	ae.hardDenoiced.Store(!h)
	return !h
}

func (ae *audioEngine) OnOffSoftDenoice() bool {
	s := ae.softDenoiced.Load()
	ae.mu.Lock()
	if ae.preprocessor != nil {
		ae.preprocessor.EnableDenoise(!s)
	}
	ae.mu.Unlock()
	ae.softDenoiced.Store(!s)
	return !s
}

func (ae *audioEngine) OnOffAEC() bool {
	s := ae.aec.Load()
	ae.aec.Store(!s)
	ae.mu.Lock()
	if s {
		ae.preprocessor.SetEchoCanceller(nil)
	} else {
		ae.preprocessor.SetEchoCanceller(ae.echoCanceller)
	}
	ae.mu.Unlock()
	return !s
}

func (ae *audioEngine) OnOffFilter() bool {
	s := ae.filtered.Load()
	ae.filtered.Store(!s)
	return !s
}

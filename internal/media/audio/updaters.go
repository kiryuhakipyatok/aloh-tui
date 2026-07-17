package audio

import (
	"aloh-tui/pkg/logger"

	"github.com/gen2brain/malgo"
)

type Updaters interface {
	UpdateMicrophones() error
	UpdateHeadphones() error
}

func (ae *audioEngine) UpdateMicrophones() error {
	microphones, err := ae.malgoCtx.Devices(malgo.Capture)
	if err != nil {
		ae.log.Error(0, "failed to get capture devices", logger.Err(err))
		return err
	}

	micsInfo := make(map[string]DeviceInfo, 0)

	for i, m := range microphones {
		di, err := ae.malgoCtx.DeviceInfo(malgo.Capture, m.ID, malgo.Shared)
		if err != nil {
			ae.log.Error(0, "failed to get capture devices info", logger.Err(err))
			return err
		}
		format := di.Formats[0]
		mi := DeviceInfo{
			Index:      i,
			Name:       m.Name(),
			SampleRate: format.SampleRate,
			Channels:   format.Channels,
		}

		switch format.Format {
		case malgo.FormatU8:
			mi.Format = int8f
		case malgo.FormatS16:
			mi.Format = int16f
		case malgo.FormatS24:
			mi.Format = int24f
		case malgo.FormatS32:
			mi.Format = int32f
		case malgo.FormatF32:
			mi.Format = float32f
		default:
			mi.Format = unk
		}
		micsInfo[m.Name()] = mi
	}
	ae.mu.Lock()
	ae.Microphones = micsInfo
	ae.mu.Unlock()
	return nil
}

func (ae *audioEngine) UpdateHeadphones() error {
	headphones, err := ae.malgoCtx.Devices(malgo.Playback)
	if err != nil {
		ae.log.Error(0, "failed to get playback devices", logger.Err(err))
		return err
	}

	headsInfo := make(map[string]DeviceInfo, 0)

	for i, m := range headphones {
		di, err := ae.malgoCtx.DeviceInfo(malgo.Playback, m.ID, malgo.Shared)
		if err != nil {
			ae.log.Error(0, "failed to get playback devices info", logger.Err(err))
			return err
		}
		format := di.Formats[0]
		hi := DeviceInfo{
			Index:      i,
			Name:       m.Name(),
			SampleRate: format.SampleRate,
			Channels:   format.Channels,
		}

		switch format.Format {
		case malgo.FormatU8:
			hi.Format = int8f
		case malgo.FormatS16:
			hi.Format = int16f
		case malgo.FormatS24:
			hi.Format = int24f
		case malgo.FormatS32:
			hi.Format = int32f
		case malgo.FormatF32:
			hi.Format = float32f
		default:
			hi.Format = unk
		}
		headsInfo[m.Name()] = hi
	}
	ae.mu.Lock()
	ae.Headphones = headsInfo
	ae.mu.Unlock()
	return nil
}

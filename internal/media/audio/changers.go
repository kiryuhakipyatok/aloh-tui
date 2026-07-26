package audio

import (
	"aloh-tui/pkg/logger"
	"unsafe"

	"github.com/gen2brain/malgo"
	"github.com/kechako/go-speexdsp"
)

type Changers interface {
	ChangeMicrophone(microphone string) error
	ChangeHeadphones(headphone string) error
}

func (ae *audioEngine) ChangeMicrophone(microphone string) error {
	ae.lifecycleMu.Lock()
	defer ae.lifecycleMu.Unlock()

	ae.switching.Store(true)
	defer ae.switching.Store(false)

	ae.captureReady.Store(false)

	var (
		micId unsafe.Pointer
		ch    uint32
	)

	if microphone != "" {
		id, err := ae.resolveDeviceByName(microphone, malgo.Capture)
		if err != nil {
			micId = nil
		} else {
			micId = id
		}
	}

	ae.mu.RLock()
	micInfo, ok := ae.Microphones[microphone]
	if ok && micId != nil {
		ch = micInfo.Channels
	}
	ae.mu.RUnlock()

	captureConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	captureConfig.Capture.Format = 0
	if ch > 2 {
		captureConfig.Capture.Channels = 2
	}
	captureConfig.SampleRate = 0
	captureConfig.Capture.DeviceID = micId

	newCaptureDevice, err := malgo.InitDevice(ae.malgoCtx.Context, captureConfig, ae.captureCallback)
	if err != nil && captureConfig.Capture.DeviceID != nil {
		captureConfig.Capture.DeviceID = nil
		newCaptureDevice, err = malgo.InitDevice(ae.malgoCtx.Context, captureConfig, ae.captureCallback)
	}
	if err != nil {
		return err
	}

	captureSampleRate := int(newCaptureDevice.SampleRate())

	newCaptureResampler, err := speexdsp.NewResampler(1, captureSampleRate, 48000, 7)
	if err != nil {
		ae.log.Error(ae.errLogCount, "failed to create new capture resampler", logger.Err(err))
		newCaptureDevice.Uninit()
		return err
	}

	ae.mu.Lock()
	oldCaptureDevice := ae.captureDevice
	oldCaptureResampler := ae.captureResampler
	oldMicrophone := ae.CurrentMicrophone

	if ok && micId != nil {
		ae.CurrentMicrophone = micInfo
	}

	ae.captureDevice = newCaptureDevice
	ae.captureResampler = newCaptureResampler

	if oldCaptureDevice != nil {
		oldCaptureDevice.Stop()
	}

	ae.voiceHolder.Store(0)
	ae.workMic = ae.workMic[:0]
	ae.micNativeBuffer = ae.micNativeBuffer[:0]
	ae.mu.Unlock()

	if err := newCaptureDevice.Start(); err != nil {
		ae.log.Error(ae.errLogCount, "failed to start new microphone, rolling back", logger.Err(err))

		newCaptureDevice.Uninit()
		newCaptureResampler.Close()

		ae.mu.Lock()
		ae.captureDevice = oldCaptureDevice
		ae.captureResampler = oldCaptureResampler
		ae.CurrentMicrophone = oldMicrophone
		ae.mu.Unlock()

		if oldCaptureDevice != nil {
			if startErr := oldCaptureDevice.Start(); startErr != nil {
				ae.log.Error(ae.errLogCount, "failed to restart old microphone after rollback", logger.Err(startErr))
			}
		}

		return err
	}

	if oldCaptureDevice != nil {
		oldCaptureDevice.Uninit()
	}
	if oldCaptureResampler != nil {
		oldCaptureResampler.Close()
	}

	return nil
}

func (ae *audioEngine) ChangeHeadphones(headphones string) error {
	ae.lifecycleMu.Lock()
	defer ae.lifecycleMu.Unlock()

	ae.switching.Store(true)
	defer ae.switching.Store(false)

	ae.playbackReady.Store(false)

	var (
		headsId unsafe.Pointer
	)

	if headphones != "" {
		id, err := ae.resolveDeviceByName(headphones, malgo.Playback)
		if err != nil {
			headsId = nil
		} else {
			headsId = id
		}
	}

	playbackConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	playbackConfig.Playback.Format = malgo.FormatS16
	playbackConfig.Playback.Channels = 1
	playbackConfig.SampleRate = 0
	playbackConfig.Playback.DeviceID = headsId

	newPlaybackDevice, err := malgo.InitDevice(ae.malgoCtx.Context, playbackConfig, ae.playbackCallback)
	if err != nil && playbackConfig.Playback.DeviceID != nil {
		playbackConfig.Playback.DeviceID = nil
		newPlaybackDevice, err = malgo.InitDevice(ae.malgoCtx.Context, playbackConfig, ae.playbackCallback)
	}
	if err != nil {
		return err
	}

	playbackSampleRate := int(newPlaybackDevice.SampleRate())

	newPlaybackResampler, err := speexdsp.NewResampler(1, playbackSampleRate, 48000, 7)
	if err != nil {
		ae.log.Error(ae.errLogCount, "failed to create new playback resampler", logger.Err(err))
		newPlaybackDevice.Uninit()
		return err
	}

	ae.mu.Lock()
	oldPlaybackDevice := ae.playbackDevice
	oldPlaybackResampler := ae.playbackResampler
	oldHeadphones := ae.CurrentHeadphones

	headsInfo, ok := ae.Headphones[headphones]
	if ok && headsId != nil {
		ae.CurrentHeadphones = headsInfo
	}

	ae.playbackDevice = newPlaybackDevice
	ae.playbackResampler = newPlaybackResampler

	if oldPlaybackDevice != nil {
		oldPlaybackDevice.Stop()
	}

	ae.playbackNativeBuffer = ae.playbackNativeBuffer[:0]
	ae.mu.Unlock()

	if err := newPlaybackDevice.Start(); err != nil {
		ae.log.Error(ae.errLogCount, "failed to start new headphones, rolling back", logger.Err(err))

		newPlaybackDevice.Uninit()
		newPlaybackResampler.Close()

		ae.mu.Lock()
		ae.playbackDevice = oldPlaybackDevice
		ae.playbackResampler = oldPlaybackResampler
		ae.CurrentHeadphones = oldHeadphones
		ae.mu.Unlock()

		if oldPlaybackDevice != nil {
			if startErr := oldPlaybackDevice.Start(); startErr != nil {
				ae.log.Error(ae.errLogCount, "failed to restart old headphones after rollback", logger.Err(startErr))
			}
		}

		return err
	}

	if oldPlaybackDevice != nil {
		oldPlaybackDevice.Uninit()
	}
	if oldPlaybackResampler != nil {
		oldPlaybackResampler.Close()
	}

	return nil
}

package video

import (
	"aloh-tui/internal/media"

	"github.com/google/uuid"
)

type Geter interface {
	GetUserWebcamFrame() string
	GetUserScreenFrame() string
	GetCurrentWebcam() media.Device
	GetUsersWebcamFramesTerminal() map[uuid.UUID]string
	GetUsersScreenFramesTerminal() map[uuid.UUID]string
}

func (ve *videoEngine) GetUserWebcamFrame() string {
	f, ok := ve.webcam.ui.frame.Load().(string)
	if !ok {
		return ""
	}
	return f
}

func (ve *videoEngine) GetUserScreenFrame() string {
	f, ok := ve.screen.ui.frame.Load().(string)
	if !ok {
		return ""
	}
	return f
}

func (ve *videoEngine) GetUsersWebcamFramesTerminal() map[uuid.UUID]string {
	ve.mu.RLock()
	defer ve.mu.RUnlock()
	var (
		f  string
		ok bool
	)
	frames := make(map[uuid.UUID]string, len(ve.usersDevices))
	for i, uv := range ve.usersDevices {
		if uv.webcam != nil {
			f, ok = uv.webcam.ui.frame.Load().(string)
			if ok && f != "" {
				frames[i] = f
			}
		}

	}

	return frames
}

func (ve *videoEngine) GetUsersScreenFramesTerminal() map[uuid.UUID]string {
	ve.mu.RLock()
	defer ve.mu.RUnlock()
	var (
		f  string
		ok bool
	)
	frames := make(map[uuid.UUID]string, len(ve.usersDevices))
	for i, uv := range ve.usersDevices {
		if uv.screen != nil {
			f, ok = uv.screen.ui.frame.Load().(string)
			if ok && f != "" {
				frames[i] = f
			}
		}

	}

	return frames
}

func (ve *videoEngine) GetCurrentWebcam() media.Device {
	ve.mu.RLock()
	defer ve.mu.RUnlock()
	return ve.webcam.current
}

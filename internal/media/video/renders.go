package video

import (
	"aloh-tui/pkg/logger"

	"github.com/google/uuid"
)

type Render interface {
	RenderUsersWebcam(id uuid.UUID, data []byte)
	RenderUsersScreen(id uuid.UUID, data []byte)
}

func (ve *videoEngine) RenderUsersWebcam(id uuid.UUID, data []byte) {
	ve.renderUsersDevice(WEBCAM, id, data)
}

func (ve *videoEngine) RenderUsersScreen(id uuid.UUID, data []byte) {
	ve.renderUsersDevice(SCREEN, id, data)
}

func (ve *videoEngine) renderUsersDevice(typee uint, id uuid.UUID, data []byte) {
	ve.mu.RLock()
	ud, ok := ve.usersDevices[id]
	ve.mu.RUnlock()
	if !ok {
		var err error
		ud, err = newUserDevices()
		if err != nil {
			ve.log.Error("failed to create new user devices", logger.Err(err))
			return
		}
		ve.mu.Lock()
		ve.usersDevices[id] = ud
		ve.mu.Unlock()
	}

	var di *deviceInfo

	switch typee {
	case WEBCAM:
		di = ud.webcam
	case SCREEN:
		di = ud.screen
	}

	if di.started.Load() {
		ve.mu.RLock()
		n := len(ve.usersDevices)
		ve.mu.RUnlock()

		di.mu.Lock()
		if di.decoderBuffer == nil || di.vp8Decoder == nil {
			di.mu.Unlock()
			return
		}
		_, err := di.decoderBuffer.Write(data)
		if err != nil {
			ve.log.Error("failed to write data", logger.Err(err))
			di.mu.Unlock()
			return
		}
		img, release, err := di.vp8Decoder.Read()
		if err != nil {
			if err.Error() == "decode failed: 5" {
				if rerr := ve.receiveKeyFrame(typee); rerr != nil {
					ve.log.Error("failed to receiveWebcamKeyFrame", logger.Err(rerr))
				}
			} else {
				ve.log.Error("failed to decode img", logger.Err(err))
			}
			di.mu.Unlock()
			return
		}
		di.mu.Unlock()
		di.ui.termFrameCount.Add(1)
		frameCount := di.ui.termFrameCount.Load()
		if ve.onWebcamTab.Load() && (frameCount%3 == 0 || frameCount <= 2) {

			var s bool
			switch typee {
			case WEBCAM:
				s = ve.webcam.started.Load()
			case SCREEN:
				s = ve.screen.started.Load()
			}
			if s {
				n++
			}

			webcamFrame, err := renderTerminalImg(float64(n), img)
			if err != nil {
				ve.log.Error("failed to render img", logger.Err(err))
				di.ui.termFrameCount.Add(-1)
				return
			}
			if di.started.Load() {
				di.ui.frame.Store(webcamFrame)
			}

		}

		di.mu.Lock()
		if di.ui.windowed.Load() && di.ui.window != nil {
			di.ui.img = img
			di.ui.window.Invalidate()
		}
		release()
		di.decoderBuffer.Reset()
		di.mu.Unlock()
	}

}

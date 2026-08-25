package video

import (
	"aloh-tui/internal/networking"
	"aloh-tui/pkg/logger"

	"gioui.org/io/system"
)

type Setter interface {
	SetNetworking(netw networking.Networking)
	SetConnected()
	SetDisconnected()
}

func (ve *videoEngine) SetDisconnected() {
	ve.connected.Store(false)
	var waitWebcam, waitScreen chan struct{}

	ve.webcam.mu.Lock()
	if ve.webcam.ui.windowed.Load() && ve.webcam.ui.window != nil {
		ve.webcam.ui.window.Perform(system.ActionClose)
	}
	ve.webcam.mu.Unlock()
	ve.screen.mu.Lock()
	if ve.screen.ui.windowed.Load() && ve.screen.ui.window != nil {
		ve.screen.ui.window.Perform(system.ActionClose)
	}
	ve.screen.mu.Unlock()

	for _, ud := range ve.usersDevices {
		ud.webcam.mu.Lock()
		if ud.webcam.ui.window != nil && ud.webcam.ui.windowed.Load() {
			ud.webcam.ui.window.Perform(system.ActionClose)
		}

		if ud.webcam.vp8Decoder != nil {
			if err := ud.webcam.vp8Decoder.Close(); err != nil {
				ve.log.Error("failed to clos webcam vp8Decoder", logger.Err(err))
			}
			ud.webcam.vp8Decoder = nil
		}
		ud.webcam.mu.Unlock()

		ud.screen.mu.Lock()
		if ud.screen.ui.window != nil && ud.screen.ui.windowed.Load() {
			ud.screen.ui.window.Perform(system.ActionClose)
		}

		if ud.screen.vp8Decoder != nil {
			if err := ud.screen.vp8Decoder.Close(); err != nil {
				ve.log.Error("failed to close screen vp8Decoder", logger.Err(err))
			}
			ud.screen.vp8Decoder = nil
		}

		ud.screen.mu.Unlock()

	}

	clear(ve.usersDevices)

	if ve.webcam.started.Load() {
		ve.webcam.started.Store(false)
		if ve.webcam.stopProcessChan != nil {
			close(ve.webcam.stopProcessChan)
			waitWebcam = ve.webcam.waitProcessChan

		}
	}

	if ve.screen.started.Load() {
		ve.screen.started.Store(false)
		if ve.screen.stopProcessChan != nil {
			close(ve.screen.stopProcessChan)
			waitScreen = ve.screen.waitProcessChan

		}
	}

	if waitWebcam != nil {
		<-waitWebcam
	}
	if waitScreen != nil {
		<-waitScreen
	}

	if err := ve.offDevice(ve.webcam); err != nil {
		ve.log.Error("failed to off webcam", logger.Err(err))
	}
	ve.webcam = nil

	if err := ve.offDevice(ve.screen); err != nil {
		ve.log.Error("failed to off screen", logger.Err(err))
	}
	ve.screen = nil

}

func (ve *videoEngine) SetNetworking(netw networking.Networking) {
	ve.mu.Lock()
	defer ve.mu.Unlock()
	ve.netw = netw
}

func (ve *videoEngine) SetConnected() {
	ve.connected.Store(true)
}

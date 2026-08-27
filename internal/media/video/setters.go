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

	if ve.webcam != nil {
		var waitWebcam chan struct{}
		ve.webcam.mu.Lock()
		if ve.webcam.ui.windowed.Load() && ve.webcam.ui.window != nil {
			ve.webcam.ui.window.Perform(system.ActionClose)
		}

		if ve.webcam.started.Load() {
			ve.webcam.started.Store(false)
			if ve.webcam.stopProcessChan != nil {
				close(ve.webcam.stopProcessChan)

				waitWebcam = ve.webcam.waitProcessChan

			}
		}
		ve.webcam.mu.Unlock()

		if waitWebcam != nil {
			<-waitWebcam
		}
		ve.webcam.stopProcessChan = nil
		ve.webcam.waitProcessChan = nil

		if err := ve.offDevice(ve.webcam); err != nil {
			ve.log.Error("failed to off webcam", logger.Err(err))
		}
		//ve.webcam = nil
	}

	if ve.screen != nil {
		var waitScreen chan struct{}

		ve.screen.mu.Lock()
		if ve.screen.ui.windowed.Load() && ve.screen.ui.window != nil {
			ve.screen.ui.window.Perform(system.ActionClose)
		}
		if ve.screen.started.Load() {
			ve.screen.started.Store(false)
			if ve.screen.stopProcessChan != nil {
				close(ve.screen.stopProcessChan)
				waitScreen = ve.screen.waitProcessChan

			}
		}
		ve.screen.mu.Unlock()
		if waitScreen != nil {
			<-waitScreen
		}
		ve.screen.stopProcessChan = nil
		ve.screen.waitProcessChan = nil
		if err := ve.offDevice(ve.screen); err != nil {
			ve.log.Error("failed to off screen", logger.Err(err))
		}
		//ve.screen = nil
	}

	for _, ud := range ve.usersDevices {
		if ud.webcam != nil {
			ud.webcam.mu.Lock()
			if ud.webcam.ui.window != nil && ud.webcam.ui.windowed.Load() {
				ud.webcam.ui.window.Perform(system.ActionClose)
			}

			if ud.webcam.h264Decoder != nil {
				if err := ud.webcam.h264Decoder.Close(); err != nil {
					ve.log.Error("failed to clos webcam h264Decoder", logger.Err(err))
				}
				ud.webcam.h264Decoder = nil
			}
			ud.webcam.mu.Unlock()

			ud.webcam = nil
		}

		if ud.screen != nil {
			ud.screen.mu.Lock()
			if ud.screen.ui.window != nil && ud.screen.ui.windowed.Load() {
				ud.screen.ui.window.Perform(system.ActionClose)
			}

			if ud.screen.h264Decoder != nil {
				if err := ud.screen.h264Decoder.Close(); err != nil {
					ve.log.Error("failed to close screen h264Decoder", logger.Err(err))
				}
				ud.screen.h264Decoder = nil
			}

			ud.screen.mu.Unlock()

			ud.screen = nil
		}

	}

	clear(ve.usersDevices)

}

func (ve *videoEngine) SetNetworking(netw networking.Networking) {
	ve.mu.Lock()
	defer ve.mu.Unlock()
	ve.netw = netw
}

func (ve *videoEngine) SetConnected() {
	ve.connected.Store(true)
}

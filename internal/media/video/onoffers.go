package video

import (
	"aloh-tui/pkg/errs"
	"aloh-tui/pkg/logger"
	"bytes"
	"fmt"
	"image"

	"gioui.org/app"
	"gioui.org/io/system"
	"github.com/google/uuid"
	"github.com/pion/mediadevices/pkg/codec/openh264"
	"github.com/pion/mediadevices/pkg/prop"
)

type OnOffer interface {
	OnOffUsersWebcamWindow(id uuid.UUID, nickname string) (bool, error)
	OnOffUsersScreenWindow(id uuid.UUID, nickname string) (bool, error)

	OnOffUsersWebcam(res bool, id uuid.UUID) error
	OnOffUsersScreen(res bool, id uuid.UUID) error

	OnOffWebcamWindow() (bool, error)
	OnOffScreenWindow() (bool, error)

	OnOffWebcam() (bool, error)
	OnOffScreen() (bool, error)
}

func (ve *videoEngine) OnOffUsersWebcamWindow(id uuid.UUID, nickname string) (bool, error) {
	return ve.onOffUsersWindow(WEBCAM, id, nickname)
}

func (ve *videoEngine) OnOffUsersScreenWindow(id uuid.UUID, nickname string) (bool, error) {
	return ve.onOffUsersWindow(SCREEN, id, nickname)
}

func (ve *videoEngine) OnOffUsersWebcam(res bool, id uuid.UUID) error {
	return ve.onOffUsersDevice(WEBCAM, res, id)
}

func (ve *videoEngine) OnOffUsersScreen(res bool, id uuid.UUID) error {
	return ve.onOffUsersDevice(SCREEN, res, id)
}

func (ve *videoEngine) OnOffWebcam() (bool, error) {
	return ve.onOffDevice(WEBCAM)
}

func (ve *videoEngine) OnOffScreen() (bool, error) {
	return ve.onOffDevice(SCREEN)
}

func (ve *videoEngine) OnOffWebcamWindow() (bool, error) {
	return ve.onOffWindow(WEBCAM)
}

func (ve *videoEngine) OnOffScreenWindow() (bool, error) {
	return ve.onOffWindow(SCREEN)
}

func (ve *videoEngine) onOffUsersDevice(typee uint, res bool, id uuid.UUID) error {
	ve.mu.RLock()
	ud, ok := ve.usersDevices[id]
	ve.mu.RUnlock()
	if !ok {
		var err error
		ud, err = newUserDevices()
		if err != nil {
			ve.log.Error("failed to create new user devices", logger.Err(err))
			return err
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

	if !res {
		di.ui.mu.Lock()
		if di.ui.windowed.Load() && di.ui.window != nil {
			di.ui.window.Perform(system.ActionClose)
		}
		di.ui.img = nil
		di.ui.frame.Store("")
		di.ui.mu.Unlock()

		di.mu.Lock()
		if err := di.h264Decoder.Close(); err != nil {
			ve.log.Error("failed to close webcam user decoder", logger.Err(err))
		}
		di.decoderBuffer = nil
		di.h264Decoder = nil
		di.mu.Unlock()

	} else {
		grow := setupGrow(typee)
		db := new(bytes.Buffer)
		db.Grow(grow)
		wDecParams, err := getDecParams(WEBCAM)
		if err != nil {
			return err
		}

		var (
			w int
			h int
		)

		switch typee {
		case WEBCAM:
			w = 1280
			h = 720
		case SCREEN:
			w = 1920
			h = 1080
		}
		propM := prop.Media{
			Video: prop.Video{
				Width:     w,
				Height:    h,
				FrameRate: 10,
			},
		}

		wH264Decoder, err := openh264.NewDecoder(db, propM, wDecParams)
		if err != nil {
			db = nil
			return err
		}
		if err != nil {
			ve.log.Error("failed to create decoder", logger.Err(err))
			return err
		}
		di.mu.Lock()
		di.decoderBuffer = db
		di.h264Decoder = wH264Decoder
		di.mu.Unlock()
	}

	di.started.Store(res)
	di.mu.Lock()
	di.ui.resizedFrame = image.NewNRGBA(image.Rect(0, 0, 160, 80))
	di.mu.Unlock()
	return nil
}

func (ve *videoEngine) onOffWindow(typee uint) (bool, error) {
	var vd *videoDevice

	switch typee {
	case WEBCAM:
		vd = ve.webcam
	case SCREEN:
		vd = ve.screen
	}

	if !vd.ui.windowed.Load() {
		window := new(app.Window)
		var title string
		switch typee {
		case WEBCAM:
			title = "your webcam"
		case SCREEN:
			title = "your screen"
		}
		window.Option(app.Title(title))
		vd.ui.mu.Lock()
		vd.ui.window = window
		vd.ui.windowed.Store(true)
		vd.ui.mu.Unlock()
		go vd.ui.proccessWindow()
		return true, nil
	}
	vd.ui.mu.Lock()
	if vd.ui.window != nil {
		vd.ui.window.Perform(system.ActionClose)
	}
	vd.ui.img = nil
	vd.ui.mu.Unlock()

	return false, nil
}

func (ve *videoEngine) onOffUsersWindow(typee uint, id uuid.UUID, nickname string) (bool, error) {
	ve.mu.RLock()
	ud, ok := ve.usersDevices[id]
	ve.mu.RUnlock()
	if !ok {
		return false, errs.ErrNotFound()
	}

	var di *deviceInfo

	switch typee {
	case WEBCAM:
		di = ud.webcam
	case SCREEN:
		di = ud.screen
	}

	if !di.ui.windowed.Load() {
		window := new(app.Window)
		window.Option(app.Title(fmt.Sprintf("%s's webcam", nickname)))
		di.ui.mu.Lock()
		di.ui.window = window
		di.ui.mu.Unlock()
		di.ui.windowed.Store(true)
		go di.ui.proccessWindow()
		return true, nil
	}
	di.ui.mu.Lock()
	if di.ui.window != nil {
		di.ui.window.Perform(system.ActionClose)
	}
	di.ui.img = nil
	di.ui.mu.Unlock()

	return false, nil

}

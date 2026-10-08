package video

import (
	"aloh-tui/internal/media"
	"aloh-tui/internal/networking"
	"aloh-tui/pkg/logger"
	"strings"
	"sync"

	"github.com/google/uuid"
	alohnetwork "github.com/kiryuhakipyatok/aloh-networking"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/driver"
	_ "github.com/pion/mediadevices/pkg/driver/camera"
	_ "github.com/pion/mediadevices/pkg/driver/screen"
)

type VideoEngine interface {
	FetchWebcams() map[string]media.Device
	ChangeWebcam(webcam string) error
	UpdateWebcams()

	Setter
	Switcher
	Render
	Sender
	OnOffer
	Geter
	Checker

	Stop()
}

type videoEngine struct {
	netw         networking.Networking
	usersDevices map[uuid.UUID]*userDevices

	webcam *videoDevice
	screen *videoDevice

	webcams map[string]media.Device

	mu  sync.RWMutex
	log *logger.Logger

	stopSendChan chan struct{}

	atomics
}

func NewVideoEngine(l *logger.Logger, vs VideoSetup) (VideoEngine, error) {
	log := l.AddOp("videoEngine")

	ve := &videoEngine{
		log: log,

		webcam: NewWebcam(),
		screen: NewScreen(),

		stopSendChan: make(chan struct{}, 1),

		usersDevices: make(map[uuid.UUID]*userDevices, 3),
	}

	webcams := ve.fetchWebcams(vs.Webcam)

	log.Info("webcams", webcams)
	ve.webcams = webcams

	go ve.sendVideo(WEBCAM)
	go ve.sendVideo(SCREEN)

	return ve, nil
}

func (ve *videoEngine) UpdateWebcams() {
	devices := mediadevices.EnumerateDevices()
	webcams := make(map[string]media.Device, len(devices))
	var name string
	for i, d := range devices {
		if d.DeviceType == driver.Camera && d.Kind == mediadevices.VideoInput {
			splited := strings.Split(d.Label, ";")
			if len(splited) > 0 {
				name = splited[0]
			} else {
				name = d.Label
			}
			di := DeviceInfo{
				Name:  name,
				Index: i,
				Id:    d.DeviceID,
			}
			webcams[d.DeviceID] = di

		}
	}
	ve.mu.Lock()
	ve.webcams = webcams
	ve.mu.Unlock()
}

func (ve *videoEngine) FetchWebcams() map[string]media.Device {
	ve.mu.RLock()
	webcs := ve.webcams
	ve.mu.RUnlock()
	return webcs
}

func (ve *videoEngine) Stop() {

	close(ve.stopSendChan)

	if ve.webcam != nil {
		if err := ve.offDevice(ve.webcam); err != nil {
			ve.log.Error("failed to off webcam", logger.Err(err))
		}
		ve.webcam = nil
	}

	if ve.screen != nil {
		if err := ve.offDevice(ve.screen); err != nil {
			ve.log.Error("failed to off screen", logger.Err(err))
		}
		ve.screen = nil
	}

}

func (ve *videoEngine) receiveKeyFrame(typee uint) error {
	ve.mu.RLock()
	netw := ve.netw
	ve.mu.RUnlock()
	if ve.connected.Load() && netw != nil {
		var (
			kfe alohnetwork.Event
			err error
		)
		switch typee {
		case WEBCAM:
			kfe, err = networking.WebcamKeyFrameEvent()
		case SCREEN:
			kfe, err = networking.ScreenKeyFrameEvent()
		}
		if err != nil {
			ve.log.Error("failed to create key frame event", logger.Err(err))
			return err
		}
		if err := netw.NewEvent(kfe); err != nil {
			ve.log.Error("failed to send key frame event", logger.Err(err))
			return err
		}
	}
	return nil
}

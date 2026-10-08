//go:build linux

package video

import (
	"aloh-tui/internal/media"
	"aloh-tui/pkg/errs"
	"strings"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/driver"
)

func (ve *videoEngine) fetchWebcams(userWebcam string) map[string]media.Device {
	devices := mediadevices.EnumerateDevices()
	webcams := make(map[string]media.Device, len(devices))
	var i int
	var name string
	for _, d := range devices {
		if d.DeviceType == driver.Camera && d.Kind == mediadevices.VideoInput {
			splited := strings.Split(d.Label, ";")
			name = splited[0]
			di := DeviceInfo{
				Name:  name,
				Index: i,
				Id:    d.DeviceID,
				///Label: ,
			}

			webcams[name] = di

			if di.Index == 0 {
				ve.webcam.current = di
			}

			if userWebcam != "" && name == userWebcam {
				ve.webcam.current = di
			}

			i++
		}
	}
	return webcams
}

func (ve *videoEngine) resolveWebcamByName(webcam string) (DeviceInfo, error) {
	var di DeviceInfo
	devices := mediadevices.EnumerateDevices()
	for i, d := range devices {
		if d.DeviceType == driver.Camera && d.Kind == mediadevices.VideoInput && strings.HasPrefix(d.Label, webcam) {
			splited := strings.Split(d.Label, ";")
			di := DeviceInfo{
				Name:  splited[0],
				Index: i,
				Id:    d.DeviceID,
			}
			return di, nil
		}
	}
	return di, errs.ErrNotFound()
}

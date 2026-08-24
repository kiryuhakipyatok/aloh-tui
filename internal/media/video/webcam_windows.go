//go:build windows

package video

import (
	"aloh-tui/internal/media"
	"aloh-tui/pkg/errs"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/driver"
)

func (ve *videoEngine) fetchWebcams(userWebcam string) map[string]media.Device {
	devices := mediadevices.EnumerateDevices()
	webcams := make(map[string]media.Device, len(devices))
	var i int
	for _, d := range devices {
		if d.DeviceType == driver.Camera && d.Kind == mediadevices.VideoInput {
			di := DeviceInfo{
				Name:  d.Name,
				Index: i,
				Id:    d.DeviceID,
			}

			webcams[d.Name] = di

			if di.Index == 0 {
				ve.currentWebcam = di
			}

			if userWebcam != "" && d.Name == userWebcam {
				ve.currentWebcam = di
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
		if d.DeviceType == driver.Camera && d.Kind == mediadevices.VideoInput && d.Name == webcam {
			di := DeviceInfo{
				Name:  d.Name,
				Index: i,
				Id:    d.DeviceID,
			}
			return di, nil
		}
	}
	return di, errs.ErrNotFound()
}

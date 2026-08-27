package video

import "github.com/pion/mediadevices/pkg/codec/openh264"

func getEncParams(typee uint) (openh264.EncParams, error) {

	encParams, err := openh264.NewEncParams()
	if err != nil {
		return encParams, err
	}

	var (
		br  int
		ut  openh264.UsageTypeEnum
		kfi uint
		rcm openh264.RCModeEnum
	)

	switch typee {
	case WEBCAM:
		kfi, br, ut, rcm = 3000, 1_000_000, openh264.CameraVideoRealTime, openh264.RCBitrateMode
	case SCREEN:
		kfi, br, ut, rcm = 6000, 600_000, openh264.ScreenContentRealTime, openh264.RCQualityMode

	}

	encParams.BitRate = br
	encParams.IntraPeriod = kfi
	encParams.KeyFrameInterval = int(kfi)
	encParams.UsageType = ut
	encParams.RCMode = rcm
	encParams.EnableFrameSkip = true

	return encParams, nil
}

func getDecParams(typee uint) (openh264.DecParams, error) {

	decParams, err := openh264.NewDecParams()
	if err != nil {
		return decParams, err
	}

	var (
		kfi, br int
	)

	switch typee {
	case WEBCAM:
		kfi, br = 3000, 1_000_000
	case SCREEN:
		kfi, br = 6000, 600_000

	}

	decParams.BitRate = br
	decParams.KeyFrameInterval = kfi
	decParams.VideoBitstreamType = openh264.VideoBitstreamAVC

	return decParams, nil
}

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
	)

	switch typee {
	case WEBCAM:
		kfi, br, ut = 3000, 500_000, openh264.CameraVideoRealTime
	case SCREEN:
		kfi, br, ut = 4000, 300_000, openh264.ScreenContentRealTime

	}

	encParams.BitRate = br
	encParams.IntraPeriod = kfi
	encParams.KeyFrameInterval = int(kfi)
	encParams.UsageType = ut

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
		kfi, br = 3000, 500_000
	case SCREEN:
		kfi, br = 6000, 300_000

	}

	decParams.BitRate = br
	decParams.KeyFrameInterval = kfi

	return decParams, nil
}

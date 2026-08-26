package video

import (
	"aloh-tui/pkg/logger"
	"bytes"
	"image"
	"os"

	"github.com/blacktop/go-termimg"
	"github.com/charmbracelet/x/term"
	"github.com/pion/mediadevices/pkg/codec/openh264"
	"github.com/pion/mediadevices/pkg/prop"
)

func setupGrow(typee uint) int {
	var grow int
	switch typee {
	case WEBCAM:
		grow = 20000
	case SCREEN:
		grow = 150000
	}

	return grow
}

func setupLog(typee uint, log *logger.Logger) *logger.Logger {
	var op string
	switch typee {
	case 0:
		op = "processVideoDevice (webcam)"
	default:
		op = "processVideoDevice (screen)"
	}
	return log.AddOp(op)
}

func calculateGrid(n float64) (rows, cols int) {
	total := int(n)
	if total <= 3 {
		return 1, total
	}

	rows = 2

	cols = (total + 1) / 2
	return rows, cols
}

func renderTerminalImg(n float64, img image.Image) (string, error) {
	imageWidget := termimg.NewImageWidgetFromImage(img)
	defer imageWidget.Clear()
	imageWidget.SetProtocol(termimg.Halfblocks)
	if n <= 0 {
		n = 1
	}

	termW, termH, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		termW, termH = 80, 24
	}

	r, c := calculateGrid(n)
	availableW := float64(termW/c) * 1.6
	availableH := float64(termH / r)

	imageWidget.SetSize(int(availableW), int(availableH))
	textMsg, err := imageWidget.Render()
	if err != nil {
		return "", err
	}

	return textMsg, nil
}

func renderLocalImg(n float64, img image.Image) (string, error) {
	imageWidget := termimg.NewImageWidgetFromImage(img)
	defer imageWidget.Clear()
	imageWidget.SetProtocol(termimg.Halfblocks)
	//imgSize := img.Bounds().Size()
	if n <= 0 {
		n = 1
	}

	termW, termH, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		termW, termH = 80, 24
	}

	r, c := calculateGrid(n)

	availableW := float64(termW/c) * 1.6
	availableH := float64(termH / r)

	imageWidget.SetSize(int(availableW), int(availableH))

	textMsg, err := imageWidget.Render()
	if err != nil {
		return "", err
	}

	return textMsg, nil
}

func newUserDevices() (*userDevices, error) {
	ud := &userDevices{}
	wdb := new(bytes.Buffer)
	sdb := new(bytes.Buffer)
	wdb.Grow(setupGrow(WEBCAM))
	sdb.Grow(setupGrow(SCREEN))
	wdi := &deviceInfo{
		decoderBuffer: wdb,
	}
	sdi := &deviceInfo{
		decoderBuffer: sdb,
	}
	propM := prop.Media{
		Video: prop.Video{
			Width:     1280,
			Height:    720,
			FrameRate: 15,
		},
	}

	wDecParams, err := getDecParams(WEBCAM)
	if err != nil {
		return nil, err
	}

	wH264Decoder, err := openh264.NewDecoder(wdb, propM, wDecParams)
	if err != nil {
		wdb = nil
		sdb = nil
		return nil, err
	}

	sDecParams, err := getDecParams(SCREEN)
	if err != nil {
		return nil, err
	}

	sH264Decoder, err := openh264.NewDecoder(sdb, propM, sDecParams)
	if err != nil {
		wdb = nil
		sdb = nil
		return nil, err
	}

	wdi.h264Decoder = wH264Decoder
	sdi.h264Decoder = sH264Decoder

	ud.webcam = wdi
	ud.screen = sdi

	return ud, nil
}

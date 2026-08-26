package video

import (
	"bytes"
	"sync"
	"sync/atomic"

	"github.com/pion/mediadevices/pkg/codec"
)

type atomics struct {
	connected atomic.Bool

	onWebcamTab atomic.Bool
	onScreenTab atomic.Bool
}

type DeviceInfo struct {
	Name  string
	Id    string
	Index int
	Label string
}

type VideoSetup struct {
	Webcam string
}

type deviceInfo struct {
	started atomic.Bool

	decoderBuffer *bytes.Buffer
	h264Decoder    codec.VideoDecoder

	ui ui

	mu sync.Mutex
}

type userDevices struct {
	webcam *deviceInfo
	screen *deviceInfo
}

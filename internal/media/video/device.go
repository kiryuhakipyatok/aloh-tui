package video

import (
	"aloh-tui/pkg/logger"
	"bytes"
	"errors"
	"image"
	"sync"
	"sync/atomic"
	"time"

	"gioui.org/io/system"
	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/driver/camera"
	"github.com/pion/mediadevices/pkg/driver/screen"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
	"golang.org/x/image/draw"
)

const (
	WEBCAM = iota
	SCREEN
)

type videoDevice struct {
	typee uint

	device driver.VideoDevice

	current DeviceInfo

	encodedReader codec.ReadCloser
	keyFrameCtrl  codec.KeyFrameController

	bufferChan       chan *bytes.Buffer
	bytesBuffersPool sync.Pool

	stopProcessChan chan struct{}
	waitProcessChan chan struct{}

	started atomic.Bool
	mu      sync.RWMutex

	ui ui
}

func NewWebcam() *videoDevice {
	return &videoDevice{
		typee: WEBCAM,

		bufferChan: make(chan *bytes.Buffer, 1),
		bytesBuffersPool: sync.Pool{
			New: func() any {
				b := new(bytes.Buffer)
				b.Grow(50000)
				return b
			},
		},
		ui: ui{
			resizedFrame: image.NewNRGBA(image.Rect(0, 0, 160, 80)),
		},
	}
}

func NewScreen() *videoDevice {
	return &videoDevice{
		typee: SCREEN,

		bufferChan: make(chan *bytes.Buffer, 1),
		bytesBuffersPool: sync.Pool{
			New: func() any {
				b := new(bytes.Buffer)
				b.Grow(300000)
				return b
			},
		},
		ui: ui{
			resizedFrame: image.NewNRGBA(image.Rect(0, 0, 160, 80)),
		},
	}
}

func (ve *videoEngine) ChangeWebcam(newWebcam string) error {
	started := ve.webcam.started.Load()
	if started && ve.connected.Load() {
		if err := ve.offDevice(ve.webcam); err != nil {
			return err
		}
	}

	resolvedWebcam, err := ve.resolveWebcamByName(newWebcam)
	if err != nil {
		ve.log.Error("failed to resolve webcam", logger.Err(err))
		return err
	}

	ve.mu.Lock()
	ve.webcam.current = resolvedWebcam
	ve.mu.Unlock()

	if started {
		if err := ve.onDevice(ve.webcam); err != nil {
			ve.log.Error("failed to on webcam", logger.Err(err))
			return err
		}

		ve.webcam.started.Store(true)
	}

	return nil
}

func (ve *videoEngine) onOffDevice(typee uint) (bool, error) {
	log := setupLog(typee, ve.log)
	var vd *videoDevice
	switch typee {
	case WEBCAM:
		vd = ve.webcam
	case SCREEN:
		vd = ve.screen
	}

	s := vd.started.Load()

	if !s == true {
		if err := ve.onDevice(vd); err != nil {
			log.Info("failed to on device")
			return false, err
		}
	} else {
		if err := ve.offDevice(vd); err != nil {
			log.Info("failed to off device")
			return false, err
		}
	}
	vd.started.Store(!s)
	vd.ui.termFrameCount.Store(0)
	vd.ui.resizedFrame = image.NewNRGBA(image.Rect(0, 0, 160, 80))
	return !s, nil
}

func (vd *videoDevice) processVideoDevice(wc chan struct{}, log *logger.Logger) {

	log = setupLog(vd.typee, log)
	defer func() {
		if vd.waitProcessChan != nil {
			close(vd.waitProcessChan)
		}
	}()

	closeWaitChan := sync.OnceFunc(func() {
		close(wc)
	})

	for {
		select {
		case <-vd.stopProcessChan:
			return
		default:
			vd.mu.RLock()
			ecodedReader := vd.encodedReader
			vd.mu.RUnlock()
			if ecodedReader == nil {
				return
			}
			now := time.Now()
			encodedFrame, realese, err := ecodedReader.Read()
			if err != nil {
				log.Error("failed to read encodedFrame", logger.Err(err))
				continue
			}

			log.Info("frame time", time.Since(now).Milliseconds())

			log.Info("len", len(encodedFrame))
			buffer := vd.bytesBuffersPool.Get().(*bytes.Buffer)

			_, err = buffer.Write(encodedFrame)
			if err != nil {
				log.Error("failed to write in buffer", logger.Err(err))
				realese()
				continue
			}
			vd.ui.termFrameCount.Add(1)
			realese()

			select {
			case vd.bufferChan <- buffer:
				closeWaitChan()

			default:

				buffer.Reset()
				vd.bytesBuffersPool.Put(buffer)
			}
		}

	}

}

func (ve *videoEngine) onDevice(vd *videoDevice) error {
	log := setupLog(vd.typee, ve.log)
	vd.mu.RLock()
	curW := vd.current
	vd.mu.RUnlock()

	encParams, err := getEncParams(vd.typee)
	if err != nil {
		return err
	}

	var (
		device driver.VideoDevice
		propM  prop.Video
	)
	switch vd.typee {
	case WEBCAM:
		cameraNames, err := camera.GetCameraNames()
		if err != nil {
			log.Error("failed to get cameras names", logger.Err(err))
			return err
		}
		name, ok := cameraNames[curW.Name]
		if !ok {
			ic := "invalid camera"
			log.Error(ic)
			return errors.New(ic)
		}
		camera := camera.NewCamera(name)
		if err := camera.Open(); err != nil {
			log.Error("failed to open camera", logger.Err(err))
			return err
		}
		propM = prop.Video{
			Width:       640,
			Height:      480,
			FrameRate:   30,
			FrameFormat: frame.FormatYUYV,
		}
		device = camera

	case SCREEN:
		screen := screen.NewScreen(0)
		if err := screen.Open(); err != nil {
			log.Error("failed to open screen", logger.Err(err))
			return err
		}
		log.Info("capture name", logger.Attr("name", screen.GetCaptureName()))
		propM = prop.Video{
			Width:       1920,
			Height:      1080,
			FrameRate:   30,
			FrameFormat: frame.FormatYUYV,
		}
		device = screen
	}

	reader, err := device.VideoRecord(prop.Media{
		Video: propM,
	})
	if err != nil {
		log.Error("failed to create video reader", logger.Err(err))
		return err
	}
	interceptor := video.ReaderFunc(func() (img image.Image, release func(), err error) {
		img, release, err = reader.Read()
		if err != nil {
			return nil, nil, err
		}
		ve.log.Info("img bounds", img.Bounds())
		frameN := vd.ui.termFrameCount.Load()
		ve.mu.RLock()
		n := len(ve.usersDevices)
		rf := vd.ui.resizedFrame
		ve.mu.RUnlock()

		var tab bool
		switch vd.typee {
		case WEBCAM:
			tab = ve.onWebcamTab.Load()
		case SCREEN:
			tab = ve.onScreenTab.Load()
		}

		if tab && (frameN%5 == 0 || frameN <= 1) {
			var (
				f string
				o string
			)

			switch vd.typee {
			case WEBCAM:
				o = ve.GetUserWebcamFrame()
			case SCREEN:
				o = ve.GetUserScreenFrame()
			}

			if rf != nil {
				draw.NearestNeighbor.Scale(rf, rf.Rect, img,
					img.Bounds(), draw.Over, nil)
				f, err = renderLocalImg(float64(n+1), rf)
				if err != nil {
					f = o
				}

				vd.ui.frame.Store(f)
			}

		}

		vd.mu.Lock()
		if vd.ui.windowed.Load() && vd.ui.window != nil {
			vd.ui.img = img
			vd.ui.window.Invalidate()
		}
		vd.mu.Unlock()

		return
	})

	encodedReader, err := encParams.BuildVideoEncoder(interceptor, prop.Media{
		Video: propM,
	})

	ctrl := encodedReader.Controller()

	fCtrl, ok := ctrl.(codec.KeyFrameController)
	if !ok {
		fc := "failed to cast"
		log.Error(fc)
		return errors.New(fc)
	}
	vd.encodedReader = encodedReader

	vd.device = device
	vd.keyFrameCtrl = fCtrl

	vd.stopProcessChan = make(chan struct{}, 1)
	vd.waitProcessChan = make(chan struct{}, 1)
	waitChan := make(chan struct{}, 1)

	go vd.processVideoDevice(waitChan, ve.log)
	<-waitChan
	log.Info("device started")
	return nil
}

func (ve *videoEngine) offDevice(vd *videoDevice) error {
	log := setupLog(vd.typee, ve.log)
	vd.ui.mu.Lock()
	if vd.ui.windowed.Load() && vd.ui.window != nil {
		vd.ui.window.Perform(system.ActionClose)
	}
	vd.ui.mu.Unlock()

	var wait chan struct{}
	vd.mu.Lock()
	if vd.stopProcessChan != nil {
		close(vd.stopProcessChan)
		if vd.waitProcessChan != nil {
			wait = vd.waitProcessChan

		}
	}
	vd.mu.Unlock()

	if wait != nil {

		<-wait
	}

	vd.mu.Lock()
	vd.waitProcessChan = nil
	vd.stopProcessChan = nil
	if vd.device != nil {
		if err := vd.device.Close(); err != nil {
			log.Error("failed to close device", logger.Err(err))
		}
		vd.device = nil
	}

	if vd.encodedReader != nil {
		if err := vd.encodedReader.Close(); err != nil {
			log.Error("failed to close encodedReader", logger.Err(err))
		}
		vd.encodedReader = nil
	}

	vd.keyFrameCtrl = nil
	grow := setupGrow(vd.typee)
	vd.bytesBuffersPool = sync.Pool{
		New: func() any {
			b := new(bytes.Buffer)
			b.Grow(grow)
			return b
		},
	}

	vd.mu.Unlock()

	vd.ui.img = nil
	vd.ui.frame.Store("")
	vd.ui.termFrameCount.Store(0)
	return nil
}

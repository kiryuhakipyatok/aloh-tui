package video

import (
	"aloh-tui/internal/media"
	"aloh-tui/internal/networking"
	"aloh-tui/pkg/errs"
	"aloh-tui/pkg/logger"
	"bytes"
	"fmt"
	"image"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/image/draw"
	"golang.org/x/term"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/io/system"
	"gioui.org/op"
	"gioui.org/op/paint"
	"github.com/blacktop/go-termimg"
	"github.com/google/uuid"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/codec/vpx"
	"github.com/pion/mediadevices/pkg/driver"
	_ "github.com/pion/mediadevices/pkg/driver/camera"
	"github.com/pion/mediadevices/pkg/driver/screen"
	_ "github.com/pion/mediadevices/pkg/driver/screen"
	"github.com/pion/mediadevices/pkg/frame"

	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

type VideoEngine interface {
	SetNetworking(netw networking.Networking)
	SetConnected()
	SetDisconnected()
	RenderUsersWebcam(id uuid.UUID, data []byte)
	RenderUsersScreen(id uuid.UUID, data []byte)
	OnOffUsersWebcamWindow(id uuid.UUID, nickname string) (bool, error)
	OnOffUsersScreenWindow(id uuid.UUID, nickname string) (bool, error)
	GetUsersWebcamFramesTerminal() map[uuid.UUID]string
	GetUsersScreenFramesTerminal() map[uuid.UUID]string
	FetchWebcams() map[string]media.Device
	GetCurrentWebcam() media.Device
	OnOffUsersScreen(res bool, id uuid.UUID)
	OnOffUsersWebcam(res bool, id uuid.UUID)
	OnOffWebcamWindow() (bool, error)
	OnOffScreenWindow() (bool, error)
	SwitchOnWebcamTab(res bool)
	SendWebcamKeyFrame() error
	SendScreenKeyFrame() error
	ChangeWebcam(webcam string) error
	UpdateWebcams()
	SwitchOnScreenTab(res bool)
	OnOffWebcam() (bool, error)
	OnOffScreen() (bool, error)
	GetUserWebcamFrame() string
	GetUserScreenFrame() string
	IsWebcamStarted() bool
	IsScreenStarted() bool
	Stop()
}

type videoEngine struct {
	//webcamReader        video.Reader
	encodedWebcamReader codec.ReadCloser
	webcamTrack         *mediadevices.VideoTrack
	webcamKeyFrameCtrl  codec.KeyFrameController

	screen              *screen.Screen
	encodedScreenReader codec.ReadCloser
	screenKeyFrameCtrl  codec.KeyFrameController

	netw networking.Networking

	webcamBufferChan chan *bytes.Buffer
	screenBufferChan chan *bytes.Buffer

	stopSendVideoChan     chan struct{}
	stopProcessWebcamChan chan struct{}
	waitProcessWebcamChan chan struct{}
	stopProcessScreenChan chan struct{}
	waitProcessScreenChan chan struct{}

	webcamBytesBuffersPool sync.Pool
	screenBytesBuffersPool sync.Pool

	currentWebcam media.Device
	webcams       map[string]media.Device

	connected atomic.Bool

	onWebcamTab atomic.Bool
	onScreenTab atomic.Bool

	usersVideo map[uuid.UUID]*userVideo

	userVideo *userVideo

	mu sync.RWMutex

	log *logger.Logger
}

type DeviceInfo struct {
	Name  string
	Index int
}

type VideoSetup struct {
	Webcam string
}

type userVideo struct {
	webcamStarted  atomic.Bool
	screenStarted  atomic.Bool
	webcamFrame    atomic.Value
	webcamImg      image.Image
	mu             sync.RWMutex
	webcamWindow   *app.Window
	windowedWebcam atomic.Bool
	screenFrame    atomic.Value
	screenImg      image.Image
	screenWindow   *app.Window
	windowedScreen atomic.Bool

	termWebcamFrameCount atomic.Int32
	termScreenFrameCount atomic.Int32

	webcamDecoderBuffer *bytes.Buffer
	webcamVp8Decoder    codec.VideoDecoder

	screenDecoderBuffer *bytes.Buffer
	screenVp8Decoder    codec.VideoDecoder

	resizedScreenFrame *image.NRGBA
	resizedWebcamFrame *image.NRGBA
}

func NewVideoEngine(l *logger.Logger, vs VideoSetup) (VideoEngine, error) {
	log := l.AddOp("videoEngine")

	webcamBuffer := new(bytes.Buffer)
	webcamBuffer.Grow(20000)

	screenBuffer := new(bytes.Buffer)
	screenBuffer.Grow(150000)

	ve := &videoEngine{
		log:               log,
		webcamBufferChan:  make(chan *bytes.Buffer, 1),
		screenBufferChan:  make(chan *bytes.Buffer, 1),
		stopSendVideoChan: make(chan struct{}, 1),
		usersVideo:        make(map[uuid.UUID]*userVideo, 3),
		userVideo:         &userVideo{},

		webcamBytesBuffersPool: sync.Pool{
			New: func() any {
				b := new(bytes.Buffer)
				b.Grow(20000)
				return b
			},
		},
		screenBytesBuffersPool: sync.Pool{
			New: func() any {
				b := new(bytes.Buffer)
				b.Grow(150000)
				return b
			},
		},
	}

	devices := mediadevices.EnumerateDevices()
	webcams := make(map[string]media.Device, len(devices))
	var i int
	for _, d := range devices {
		if d.DeviceType == driver.Camera && d.Kind == mediadevices.VideoInput {
			di := DeviceInfo{
				Name:  d.Label,
				Index: i,
			}
			webcams[d.Label] = di

			if vs.Webcam != "" && vs.Webcam == d.Label {
				ve.currentWebcam = di
				continue
			}

			ve.currentWebcam = di
			i++
		}
	}

	ve.log.Info("webcams", webcams)
	ve.webcams = webcams

	go ve.sendWebcam()
	go ve.sendScreen()

	return ve, nil
}

func (ve *videoEngine) UpdateWebcams() {
	devices := mediadevices.EnumerateDevices()
	webcams := make(map[string]media.Device, len(devices))

	for i, d := range devices {
		if d.DeviceType == driver.Camera && d.Kind == mediadevices.VideoInput {
			di := DeviceInfo{
				Name:  d.DeviceID,
				Index: i,
			}
			webcams[d.DeviceID] = di

		}
	}
	ve.mu.Lock()
	ve.webcams = webcams
	ve.mu.Unlock()
}

func (ve *videoEngine) resolveWebcamByName(webcam string) (DeviceInfo, error) {
	var di DeviceInfo
	devices := mediadevices.EnumerateDevices()
	for i, d := range devices {
		if d.DeviceType == driver.Camera && d.Kind == mediadevices.VideoInput && d.Label == webcam {
			di.Name = d.Label
			di.Index = i
			return di, nil
		}
	}
	return di, errs.ErrNotFound()
}

func (ve *videoEngine) ChangeWebcam(webcam string) error {
	started := ve.userVideo.webcamStarted.Load()
	if started && ve.connected.Load() {
		ve.mu.Lock()
		ve.userVideo.webcamImg = nil
		ve.userVideo.resizedWebcamFrame = image.NewNRGBA(image.Rect(0, 0, 160, 80))
		ve.mu.Unlock()

		ve.userVideo.webcamFrame.Store("")
		ve.userVideo.webcamStarted.Store(false)
		ve.userVideo.mu.Lock()
		if ve.userVideo.webcamWindow != nil {
			ve.userVideo.webcamWindow.Perform(system.ActionClose)
		}
		ve.userVideo.mu.Unlock()

		ve.mu.Lock()
		var waitWebcam chan struct{}
		if ve.stopProcessWebcamChan != nil {
			close(ve.stopProcessWebcamChan)
			if ve.waitProcessWebcamChan != nil {
				waitWebcam = ve.waitProcessWebcamChan

			}
		}
		ve.mu.Unlock()
		if waitWebcam != nil {
			<-waitWebcam
		}

		ve.mu.Lock()
		ve.waitProcessWebcamChan = nil
		ve.stopProcessWebcamChan = nil
		if ve.webcamTrack != nil {
			if err := ve.webcamTrack.Close(); err != nil {
				ve.log.Error("failed to close webcam track", logger.Err(err))
			}
			ve.webcamTrack = nil
		}
		if ve.encodedWebcamReader != nil {
			if err := ve.encodedWebcamReader.Close(); err != nil {
				ve.log.Error("failed to close encodedWebcamReader track", logger.Err(err))
			}
			ve.encodedWebcamReader = nil
		}
		// ve.webcamReader = nil
		ve.encodedWebcamReader = nil
		ve.webcamKeyFrameCtrl = nil
		ve.mu.Unlock()
	}

	resolvedWebcam, err := ve.resolveWebcamByName(webcam)
	if err != nil {
		ve.log.Error("failed to resolve webcam", logger.Err(err))
		return err
	}

	ve.mu.Lock()
	ve.currentWebcam = resolvedWebcam
	ve.mu.Unlock()

	if started {
		vp8Params, _ := vpx.NewVP8Params()
		vp8Params.KeyFrameInterval = 3000
		vp8Params.ErrorResilient = vpx.ErrorResilientDefault
		vp8Params.RateControlEndUsage = vpx.RateControlCBR
		vp8Params.BitRate = 700_000
		vp8Params.Deadline = 1 * time.Microsecond
		vp8Params.RateControlOvershootPercent = 15
		vp8Params.RateControlUndershootPercent = 50
		vp8Params.CPUUsed = 8
		vp8Params.RateControlMinQuantizer = 20
		vp8Params.RateControlMaxQuantizer = 63
		vp8Params.LagInFrames = 0

		codecSelector := mediadevices.NewCodecSelector(
			mediadevices.WithVideoEncoders(&vp8Params),
		)
		stream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
			Video: func(mtc *mediadevices.MediaTrackConstraints) {
				mtc.FrameFormat = prop.FrameFormat(frame.FormatYUYV)
				mtc.FrameRate = prop.Float(15)
				mtc.DeviceID = prop.String(resolvedWebcam.Name)
				mtc.Width = prop.Int(1280)
				mtc.Height = prop.Int(720)
			},
			Codec: codecSelector,
		})

		if err != nil {
			ve.log.Error("failed to get stream webcam", logger.Err(err))
			return err
		}

		track := stream.GetVideoTracks()[0]

		videoTrack := track.(*mediadevices.VideoTrack)

		r := videoTrack.NewReader(false)
		interceptor := video.ReaderFunc(func() (img image.Image, release func(), err error) {
			img, release, err = r.Read()
			if err != nil {
				return nil, nil, err
			}
			frameN := ve.userVideo.termWebcamFrameCount.Load()

			if ve.onWebcamTab.Load() && (frameN%3 == 0 || frameN <= 2) {
				var f string
				o := ve.GetUserWebcamFrame()
				ve.mu.RLock()
				n := len(ve.usersVideo)
				resizedWebcamFrame := ve.userVideo.resizedWebcamFrame
				ve.mu.RUnlock()

				if resizedWebcamFrame != nil {
					draw.NearestNeighbor.Scale(resizedWebcamFrame, resizedWebcamFrame.Rect, img,
						img.Bounds(), draw.Over, nil)
					f, err = renderLocalImg(float64(n+1), resizedWebcamFrame)
					if err != nil {
						ve.log.Error("failed to render img", logger.Err(err))
						f = o
					}
					ve.userVideo.webcamFrame.Store(f)
				}

			}

			ve.userVideo.mu.Lock()
			if ve.userVideo.windowedWebcam.Load() && ve.userVideo.webcamWindow != nil {
				ve.userVideo.webcamImg = img
				ve.userVideo.webcamWindow.Invalidate()
			}
			ve.userVideo.mu.Unlock()

			return
		})

		// videoTrack.Transform(func(r video.Reader) video.Reader {
		// 	return video.ReaderFunc(func() (img image.Image, release func(), err error) {
		// 		img, release, err = r.Read()
		// 		if err != nil {
		// 			return nil, nil, err
		// 		}
		// 		frameN := ve.userVideo.termWebcamFrameCount.Load()

		// 		if ve.onWebcamTab.Load() && (frameN%3 == 0 || frameN <= 2) {
		// 			var f string
		// 			o := ve.GetUserWebcamFrame()
		// 			ve.mu.RLock()
		// 			n := len(ve.usersVideo)
		// 			resizedWebcamFrame := ve.userVideo.resizedWebcamFrame
		// 			ve.mu.RUnlock()

		// 			if resizedWebcamFrame != nil {
		// 				draw.NearestNeighbor.Scale(resizedWebcamFrame, resizedWebcamFrame.Rect, img,
		// 					img.Bounds(), draw.Over, nil)
		// 				f, err = renderLocalImg(float64(n+1), resizedWebcamFrame)
		// 				if err != nil {
		// 					ve.log.Error("failed to render img", logger.Err(err))
		// 					f = o
		// 				}
		// 				ve.userVideo.webcamFrame.Store(f)
		// 			}

		// 		}

		// 		ve.userVideo.mu.Lock()
		// 		if ve.userVideo.windowedWebcam.Load() && ve.userVideo.webcamWindow != nil {
		// 			ve.userVideo.webcamImg = img
		// 			ve.userVideo.webcamWindow.Invalidate()
		// 		}
		// 		ve.userVideo.mu.Unlock()

		// 		return
		// 	})

		// })
		// webcamReader := videoTrack.NewReader(false)
		encodedWebcamReader, err := vp8Params.BuildVideoEncoder(interceptor, prop.Media{
			Video: prop.Video{
				Width:       1280,
				Height:      720,
				FrameRate:   15,
				FrameFormat: frame.FormatYUYV,
			},
			DeviceID: resolvedWebcam.Name,
		})

		ctrl := encodedWebcamReader.Controller()

		fCtrl, ok := ctrl.(codec.KeyFrameController)
		if !ok {
			ve.log.Error("failed to casr KeyFrameController", logger.Err(err))
			return err
		}

		ve.webcamKeyFrameCtrl = fCtrl
		//ve.webcamReader = webcamReader
		ve.encodedWebcamReader = encodedWebcamReader

		ve.stopProcessWebcamChan = make(chan struct{}, 1)
		ve.waitProcessWebcamChan = make(chan struct{}, 1)
		waitChan := make(chan struct{}, 1)
		go ve.processWebcam(waitChan)
		<-waitChan
		ve.userVideo.webcamStarted.Store(true)
	}
	return nil
}

func (ve *videoEngine) FetchWebcams() map[string]media.Device {
	ve.mu.RLock()
	webcs := ve.webcams
	ve.mu.RUnlock()
	return webcs
}

func (ve *videoEngine) GetCurrentWebcam() media.Device {
	ve.mu.RLock()
	defer ve.mu.RUnlock()
	return ve.currentWebcam
}

func (ve *videoEngine) SetNetworking(netw networking.Networking) {
	ve.mu.Lock()
	defer ve.mu.Unlock()
	ve.netw = netw
}

func (ve *videoEngine) SetConnected() {
	ve.connected.Store(true)
}

func (ve *videoEngine) IsWebcamStarted() bool {
	return ve.userVideo.webcamStarted.Load()
}

func (ve *videoEngine) IsScreenStarted() bool {
	return ve.userVideo.screenStarted.Load()
}

func (ve *videoEngine) GetUserWebcamFrame() string {
	f, ok := ve.userVideo.webcamFrame.Load().(string)
	if !ok {
		return ""
	}
	return f
}

func (ve *videoEngine) GetUserScreenFrame() string {
	f, ok := ve.userVideo.screenFrame.Load().(string)
	if !ok {
		return ""
	}
	return f
}

func (ve *videoEngine) SwitchOnWebcamTab(res bool) {
	ve.onWebcamTab.Store(res)
}

func (ve *videoEngine) SwitchOnScreenTab(res bool) {
	ve.onScreenTab.Store(res)
}

func (ve *videoEngine) OnOffWebcamWindow() (bool, error) {
	ve.mu.Lock()
	defer ve.mu.Unlock()

	if !ve.userVideo.windowedWebcam.Load() {
		window := new(app.Window)
		window.Option(app.Title("your webcam"))
		ve.userVideo.mu.Lock()
		ve.userVideo.webcamWindow = window
		ve.userVideo.windowedWebcam.Store(true)
		ve.userVideo.mu.Unlock()
		go ve.userVideo.proccessUsersWindowWebcam()
		return true, nil
	}
	ve.userVideo.mu.Lock()
	if ve.userVideo.webcamWindow != nil {
		ve.userVideo.webcamWindow.Perform(system.ActionClose)
	}
	ve.userVideo.webcamImg = nil
	ve.userVideo.mu.Unlock()

	return false, nil
}

func (ve *videoEngine) OnOffScreenWindow() (bool, error) {
	ve.mu.Lock()
	defer ve.mu.Unlock()

	if !ve.userVideo.windowedScreen.Load() {
		window := new(app.Window)
		window.Option(app.Title("your screen"))
		ve.userVideo.mu.Lock()
		ve.userVideo.screenWindow = window
		ve.userVideo.windowedScreen.Store(true)
		ve.userVideo.mu.Unlock()
		go ve.userVideo.proccessUsersWindowScreen()
		return true, nil
	}
	ve.userVideo.mu.Lock()
	if ve.userVideo.screenWindow != nil {
		ve.userVideo.screenWindow.Perform(system.ActionClose)
	}
	ve.userVideo.screenImg = nil
	ve.userVideo.mu.Unlock()

	return false, nil
}

func (ve *videoEngine) OnOffUsersWebcamWindow(id uuid.UUID, nickname string) (bool, error) {
	ve.mu.Lock()
	defer ve.mu.Unlock()

	uv, ok := ve.usersVideo[id]
	if !ok {
		return false, errs.ErrNotFound()
	}

	if !uv.windowedWebcam.Load() {
		window := new(app.Window)
		window.Option(app.Title(fmt.Sprintf("%s's webcam", nickname)))
		uv.mu.Lock()
		uv.webcamWindow = window
		uv.windowedWebcam.Store(true)
		uv.mu.Unlock()
		go uv.proccessUsersWindowWebcam()
		return true, nil
	}
	uv.mu.Lock()
	if uv.webcamWindow != nil {
		uv.webcamWindow.Perform(system.ActionClose)
	}
	uv.mu.Unlock()
	uv.webcamImg = nil
	return false, nil

}

func (ve *videoEngine) OnOffUsersScreenWindow(id uuid.UUID, nickname string) (bool, error) {
	ve.mu.Lock()
	defer ve.mu.Unlock()

	uv, ok := ve.usersVideo[id]
	if !ok {
		return false, errs.ErrNotFound()
	}

	if !uv.windowedScreen.Load() {
		window := new(app.Window)
		window.Option(app.Title(fmt.Sprintf("%s's screen", nickname)))
		uv.mu.Lock()
		uv.screenWindow = window
		uv.windowedScreen.Store(true)
		uv.mu.Unlock()
		go uv.proccessUsersWindowScreen()
		return true, nil
	}
	uv.mu.Lock()
	if uv.screenWindow != nil {
		uv.screenWindow.Perform(system.ActionClose)
	}
	uv.screenImg = nil
	uv.mu.Unlock()

	return false, nil

}

func (ve *videoEngine) OnOffWebcam() (bool, error) {
	s := ve.userVideo.webcamStarted.Load()
	ve.mu.Lock()
	ve.userVideo.resizedWebcamFrame = image.NewNRGBA(image.Rect(0, 0, 160, 80))
	ve.userVideo.webcamImg = nil
	ve.mu.Unlock()

	if !s == true {
		ve.mu.RLock()
		curW, ok := ve.currentWebcam.(DeviceInfo)
		if !ok {
			return false, errs.ErrInvalidType
		}
		ve.mu.RUnlock()

		vp8Params, err := vpx.NewVP8Params()
		if err != nil {
			return false, err
		}
		vp8Params.KeyFrameInterval = 3000
		vp8Params.ErrorResilient = vpx.ErrorResilientDefault
		vp8Params.RateControlEndUsage = vpx.RateControlCBR
		vp8Params.BitRate = 700_000
		vp8Params.Deadline = 1 * time.Microsecond
		vp8Params.RateControlOvershootPercent = 15
		vp8Params.RateControlUndershootPercent = 50
		vp8Params.CPUUsed = 8
		vp8Params.RateControlMinQuantizer = 20
		vp8Params.RateControlMaxQuantizer = 63
		vp8Params.LagInFrames = 0

		codecSelector := mediadevices.NewCodecSelector(
			mediadevices.WithVideoEncoders(&vp8Params),
		)
		stream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
			Video: func(mtc *mediadevices.MediaTrackConstraints) {
				mtc.FrameFormat = prop.FrameFormat(frame.FormatYUYV)
				mtc.DeviceID = prop.String(curW.Name)
				mtc.FrameRate = prop.Float(15)
				mtc.Width = prop.Int(1280)
				mtc.Height = prop.Int(720)
			},
			Codec: codecSelector,
		})

		if err != nil {
			ve.log.Error("failed to get stream webcam", logger.Err(err))
			return false, err
		}

		track := stream.GetVideoTracks()[0]

		videoTrack := track.(*mediadevices.VideoTrack)
		r := videoTrack.NewReader(false)
		interceptor := video.ReaderFunc(func() (img image.Image, release func(), err error) {
			img, release, err = r.Read()
			if err != nil {
				return nil, nil, err
			}
			frameN := ve.userVideo.termWebcamFrameCount.Load()

			if ve.onWebcamTab.Load() && (frameN%3 == 0 || frameN <= 2) {
				var f string
				o := ve.GetUserWebcamFrame()
				ve.mu.RLock()
				n := len(ve.usersVideo)
				resizedWebcamFrame := ve.userVideo.resizedWebcamFrame
				ve.mu.RUnlock()

				if resizedWebcamFrame != nil {
					draw.NearestNeighbor.Scale(resizedWebcamFrame, resizedWebcamFrame.Rect, img,
						img.Bounds(), draw.Over, nil)
					f, err = renderLocalImg(float64(n+1), resizedWebcamFrame)
					if err != nil {
						ve.log.Error("failed to render img", logger.Err(err))
						f = o
					}
					ve.userVideo.webcamFrame.Store(f)
				}

			}

			ve.userVideo.mu.Lock()
			if ve.userVideo.windowedWebcam.Load() && ve.userVideo.webcamWindow != nil {
				ve.userVideo.webcamImg = img
				ve.userVideo.webcamWindow.Invalidate()
			}
			ve.userVideo.mu.Unlock()

			return
		})

		// videoTrack.Transform(func(r video.Reader) video.Reader {
		// 	return video.ReaderFunc(func() (img image.Image, release func(), err error) {
		// 		img, release, err = r.Read()
		// 		if err != nil {
		// 			return nil, nil, err
		// 		}
		// 		frameN := ve.userVideo.termWebcamFrameCount.Load()

		// 		if ve.onWebcamTab.Load() && (frameN%3 == 0 || frameN <= 2) {
		// 			var f string
		// 			o := ve.GetUserWebcamFrame()
		// 			ve.mu.RLock()
		// 			n := len(ve.usersVideo)
		// 			resizedWebcamFrame := ve.userVideo.resizedWebcamFrame
		// 			ve.mu.RUnlock()

		// 			if resizedWebcamFrame != nil {
		// 				draw.NearestNeighbor.Scale(resizedWebcamFrame, resizedWebcamFrame.Rect, img,
		// 					img.Bounds(), draw.Over, nil)
		// 				f, err = renderLocalImg(float64(n+1), resizedWebcamFrame)
		// 				if err != nil {
		// 					ve.log.Error("failed to render img", logger.Err(err))
		// 					f = o
		// 				}
		// 				ve.userVideo.webcamFrame.Store(f)
		// 			}

		// 		}

		// 		ve.userVideo.mu.Lock()
		// 		if ve.userVideo.windowedWebcam.Load() && ve.userVideo.webcamWindow != nil {
		// 			ve.userVideo.webcamImg = img
		// 			ve.userVideo.webcamWindow.Invalidate()
		// 		}
		// 		ve.userVideo.mu.Unlock()

		// 		return
		// 	})

		// })

		//	webcamReader := videoTrack.NewReader(false)
		// encodedWebcamReader, err := videoTrack.NewEncodedReader("vp8")
		// if err != nil {
		// 	if terr := videoTrack.Close(); terr != nil {
		// 		ve.log.Error("failed to close screen videoTrack", logger.Err(terr))
		// 	}
		// 	videoTrack = nil
		// 	//webcamReader = nil
		// 	ve.log.Error("failed to create new encoded reader", logger.Err(err))
		// 	return false, err
		// }
		encodedWebcamReader, err := vp8Params.BuildVideoEncoder(interceptor, prop.Media{
			Video: prop.Video{
				Width:       1280,
				Height:      720,
				FrameRate:   15,
				FrameFormat: frame.FormatYUYV,
			},
			DeviceID: curW.Name,
		})

		ctrl := encodedWebcamReader.Controller()

		fCtrl, ok := ctrl.(codec.KeyFrameController)
		if !ok {
			ve.log.Error("failed to casr KeyFrameController", logger.Err(err))
			return false, err
		}

		//r := encodedWebcamReader.Controller().(codec.KeyFrameController)

		ve.encodedWebcamReader = encodedWebcamReader

		//ve.webcamReader = webcamReader
		ve.webcamTrack = videoTrack
		ve.webcamKeyFrameCtrl = fCtrl

		ve.stopProcessWebcamChan = make(chan struct{}, 1)
		ve.waitProcessWebcamChan = make(chan struct{}, 1)
		waitChan := make(chan struct{}, 1)

		go ve.processWebcam(waitChan)
		<-waitChan

	} else {
		ve.userVideo.mu.Lock()
		if ve.userVideo.windowedWebcam.Load() && ve.userVideo.webcamWindow != nil {
			ve.userVideo.webcamWindow.Perform(system.ActionClose)
		}
		ve.userVideo.mu.Unlock()

		var waitWebcam chan struct{}
		ve.mu.Lock()
		if ve.stopProcessWebcamChan != nil {
			close(ve.stopProcessWebcamChan)
			if ve.waitProcessWebcamChan != nil {
				waitWebcam = ve.waitProcessWebcamChan

			}
		}
		ve.mu.Unlock()

		if waitWebcam != nil {
			<-waitWebcam
		}

		ve.mu.Lock()
		ve.waitProcessWebcamChan = nil
		ve.stopProcessWebcamChan = nil
		if ve.webcamTrack != nil {
			if err := ve.webcamTrack.Close(); err != nil {
				ve.log.Error("failed to close webcam track", logger.Err(err))
			}
			ve.webcamTrack = nil
		}

		if ve.encodedWebcamReader != nil {
			if err := ve.encodedWebcamReader.Close(); err != nil {
				ve.log.Error("failed to close encodedWebcamReader", logger.Err(err))
			}
			ve.encodedWebcamReader = nil
		}

		//	ve.webcamReader = nil

		ve.webcamKeyFrameCtrl = nil

		ve.webcamBytesBuffersPool = sync.Pool{
			New: func() any {
				b := new(bytes.Buffer)
				b.Grow(20000)
				return b
			},
		}

		ve.mu.Unlock()
	}
	ve.userVideo.webcamStarted.Store(!s)
	ve.userVideo.webcamFrame.Store("")
	ve.userVideo.termWebcamFrameCount.Store(0)

	return !s, nil
}

func (ve *videoEngine) OnOffScreen() (bool, error) {
	s := ve.userVideo.screenStarted.Load()

	ve.mu.Lock()
	ve.userVideo.screenImg = nil
	ve.userVideo.resizedScreenFrame = image.NewNRGBA(image.Rect(0, 0, 160, 80))
	ve.mu.Unlock()
	ve.log.Info("screen state", s)
	if !s == true {
		vp8Params, _ := vpx.NewVP8Params()
		vp8Params.KeyFrameInterval = 6000
		vp8Params.ErrorResilient = vpx.ErrorResilientDefault
		vp8Params.RateControlEndUsage = vpx.RateControlCBR
		vp8Params.BitRate = 300_000
		vp8Params.CPUUsed = 8
		vp8Params.Deadline = 1 * time.Microsecond
		vp8Params.RateControlOvershootPercent = 15
		vp8Params.RateControlUndershootPercent = 100
		vp8Params.RateControlMinQuantizer = 20
		vp8Params.RateControlMaxQuantizer = 63
		vp8Params.LagInFrames = 0

		screen := screen.NewScreen(0)
		if err := screen.Open(); err != nil {
			return false, err
		}

		reader, err := screen.VideoRecord(prop.Media{
			Video: prop.Video{
				Width:       1280,
				Height:      720,
				FrameRate:   15,
				FrameFormat: frame.FormatYUYV,
			},
		})

		interceptor := video.ReaderFunc(func() (img image.Image, release func(), err error) {
			img, release, err = reader.Read()
			if err != nil {
				return nil, nil, err
			}

			frameN := ve.userVideo.termScreenFrameCount.Load()

			if ve.onScreenTab.Load() && (frameN%3 == 0 || frameN <= 2) {
				var f string
				o := ve.GetUserScreenFrame()
				ve.mu.RLock()
				n := len(ve.usersVideo)
				resizedScreenFrame := ve.userVideo.resizedScreenFrame
				ve.mu.RUnlock()

				if resizedScreenFrame != nil {
					draw.NearestNeighbor.Scale(resizedScreenFrame, resizedScreenFrame.Rect, img,
						img.Bounds(), draw.Over, nil)
					f, err = renderLocalImg(float64(n+1), resizedScreenFrame)
					if err != nil {
						ve.log.Error("failed to render img", logger.Err(err))
						f = o
					}
					ve.userVideo.screenFrame.Store(f)
				}

			}

			ve.userVideo.mu.Lock()
			if ve.userVideo.windowedScreen.Load() && ve.userVideo.screenWindow != nil {
				ve.userVideo.screenImg = img
				ve.userVideo.screenWindow.Invalidate()
			}
			ve.userVideo.mu.Unlock()

			return
		})
		encodedReader, err := vp8Params.BuildVideoEncoder(interceptor, prop.Media{
			Video: prop.Video{
				Width:       1280,
				Height:      720,
				FrameRate:   15,
				FrameFormat: frame.FormatYUYV,
			},
		})

		ctrl := encodedReader.Controller()

		fCtrl, ok := ctrl.(codec.KeyFrameController)
		if !ok {
			ve.log.Error("failed to cast KeyFrameController", logger.Err(err))
			return false, err
		}

		ve.screen = screen
		ve.encodedScreenReader = encodedReader
		ve.screenKeyFrameCtrl = fCtrl

		ve.stopProcessScreenChan = make(chan struct{}, 1)
		ve.waitProcessScreenChan = make(chan struct{}, 1)
		if ve.stopProcessScreenChan != nil {
			ve.log.Info("stop screen chan not nil")
		}
		if ve.waitProcessScreenChan != nil {
			ve.log.Info("wait screen chan not nil")
		}
		waitChan := make(chan struct{}, 1)
		go ve.processScreen(waitChan)
		<-waitChan
		ve.log.Info("wait chan screen")

	} else {
		ve.userVideo.mu.Lock()
		if ve.userVideo.windowedScreen.Load() && ve.userVideo.screenWindow != nil {
			ve.userVideo.screenWindow.Perform(system.ActionClose)
		}
		ve.userVideo.mu.Unlock()
		var waitScreen chan struct{}
		ve.mu.Lock()

		if ve.stopProcessScreenChan != nil {
			ve.log.Info("close stop process sreen chan")
			close(ve.stopProcessScreenChan)
			if ve.waitProcessScreenChan != nil {
				ve.log.Info("get wait process screen chan")
				waitScreen = ve.waitProcessScreenChan
			}
		}

		ve.mu.Unlock()

		if waitScreen != nil {
			ve.log.Info("wait screen")
			<-waitScreen
		}
		ve.log.Info("wait screen done")
		ve.mu.Lock()
		// if ve.screenTrack != nil {
		// 	ve.log.Info("closing screen track")
		// 	if err := ve.screenTrack.Close(); err != nil {
		// 		ve.log.Error("failed to close screen track", logger.Err(err))
		// 	}
		// 	ve.log.Info("screen track closed")
		// 	ve.screenTrack = nil
		// }
		if ve.screen != nil {
			if err := ve.screen.Close(); err != nil {
				ve.log.Error("failed to close screen", logger.Err(err))
			}
			ve.screen = nil
		}

		if ve.encodedScreenReader != nil {
			ve.log.Info("closing encodedScreenReader")
			if err := ve.encodedScreenReader.Close(); err != nil {
				ve.log.Error("failed to close encodedScreenReader", logger.Err(err))
			}
			ve.encodedScreenReader = nil
		}

		ve.waitProcessScreenChan = nil
		ve.stopProcessScreenChan = nil

		//ve.screenReader = nil

		ve.screenKeyFrameCtrl = nil
		ve.screenBytesBuffersPool = sync.Pool{
			New: func() any {
				b := new(bytes.Buffer)
				b.Grow(150000)
				return b
			},
		}

		ve.mu.Unlock()

	}
	ve.userVideo.screenFrame.Store("")
	ve.userVideo.screenStarted.Store(!s)
	ve.userVideo.termScreenFrameCount.Store(0)
	ve.log.Info("of off screen done")

	return !s, nil
}

func (ve *videoEngine) SetDisconnected() {
	var waitWebcam, waitScreen chan struct{}
	ve.mu.Lock()

	ve.userVideo.mu.Lock()
	if ve.userVideo.windowedWebcam.Load() && ve.userVideo.webcamWindow != nil {
		ve.userVideo.webcamWindow.Perform(system.ActionClose)
	}
	if ve.userVideo.windowedScreen.Load() && ve.userVideo.screenWindow != nil {
		ve.userVideo.screenWindow.Perform(system.ActionClose)
	}
	ve.userVideo.mu.Unlock()

	for _, uv := range ve.usersVideo {
		uv.mu.Lock()
		if uv.webcamWindow != nil {
			uv.webcamWindow.Perform(system.ActionClose)
		}

		if uv.screenWindow != nil {
			uv.screenWindow.Perform(system.ActionClose)

		}

		if uv.webcamVp8Decoder != nil {
			if err := uv.webcamVp8Decoder.Close(); err != nil {
				ve.log.Error("failed to close vp8Decoder", logger.Err(err))
			}
			uv.webcamVp8Decoder = nil
		}

		uv.mu.Unlock()
	}

	clear(ve.usersVideo)

	ve.connected.Store(false)

	if ve.userVideo.webcamStarted.Load() {
		ve.userVideo.webcamStarted.Store(false)
		if ve.stopProcessWebcamChan != nil {
			close(ve.stopProcessWebcamChan)
			waitWebcam = ve.waitProcessWebcamChan

		}
	}

	if ve.userVideo.screenStarted.Load() {
		ve.userVideo.screenStarted.Store(false)
		if ve.stopProcessScreenChan != nil {
			ve.log.Info("close stop process screen when disc")
			close(ve.stopProcessScreenChan)
			waitScreen = ve.waitProcessScreenChan

		}
	}

	ve.mu.Unlock()

	if waitWebcam != nil {
		<-waitWebcam
	}
	if waitScreen != nil {
		ve.log.Info("wait wait screen when disc")
		<-waitScreen
	}

	ve.mu.Lock()
	defer ve.mu.Unlock()
	ve.stopProcessWebcamChan = nil
	if ve.webcamTrack != nil {
		if err := ve.webcamTrack.Close(); err != nil {
			ve.log.Error("failed to close webcam track", logger.Err(err))
		}
		ve.webcamTrack = nil
	}
	//ve.webcamReader = nil
	if ve.encodedWebcamReader != nil {
		if err := ve.encodedWebcamReader.Close(); err != nil {
			ve.log.Error("failed to close encodedWebcamReader", logger.Err(err))
		}
		ve.encodedWebcamReader = nil
	}
	ve.stopProcessScreenChan = nil
	ve.webcamKeyFrameCtrl = nil
	if ve.screen != nil {
		if err := ve.screen.Close(); err != nil {
			ve.log.Error("failed to close screen", logger.Err(err))
		}
		ve.screen = nil
	}
	if ve.encodedScreenReader != nil {
		if err := ve.encodedScreenReader.Close(); err != nil {
			ve.log.Error("failed to close encodedScreenReader", logger.Err(err))
		}
		ve.encodedScreenReader = nil
	}
	ve.screenKeyFrameCtrl = nil
	//ve.screenReader = nil
	ve.userVideo.webcamImg = nil
	ve.userVideo.screenImg = nil
	ve.userVideo.webcamFrame.Store("")
	ve.userVideo.screenFrame.Store("")
	ve.userVideo.termScreenFrameCount.Store(0)
	ve.userVideo.termScreenFrameCount.Store(0)
}

func (ve *videoEngine) processWebcam(wc chan struct{}) {
	ve.log.Info("processing webcam")

	defer func() {
		if ve.waitProcessWebcamChan != nil {
			close(ve.waitProcessWebcamChan)
		}
	}()

	timer := time.NewTicker(66 * time.Millisecond)
	defer timer.Stop()

	closeWaitChan := sync.OnceFunc(func() {
		close(wc)
	})

	//resizedWebcamFrame := image.NewNRGBA(image.Rect(0, 0, 180, 80))
	for {
		select {
		case <-ve.stopProcessWebcamChan:
			return
		case <-timer.C:
			select {
			case <-ve.stopProcessWebcamChan:
				return
			default:
			}

			ve.mu.RLock()
			ecodedReader := ve.encodedWebcamReader
			//reader := ve.webcamReader
			ve.mu.RUnlock()
			if ecodedReader == nil {
				return
			}
			encodedWebcamFrame, realese, err := ecodedReader.Read()
			if err != nil {
				ve.log.Error("failed to read encodedWebcamFrame", logger.Err(err))
				continue
			}
			// webcamFrame, realese, err := reader.Read()
			// if err != nil {
			// 	ve.log.Error("failed to read webcamFrame", logger.Err(err))
			// 	if ve.userVideo.termWebcamFrameCount.Load() > 0 {
			// 		ve.userVideo.termWebcamFrameCount.Add(-1)
			// 	}
			// 	eRealese()
			// 	continue
			// }
			// ve.webcamBuffer.Reset()
			//buf := ve.webcamBuffer.Bytes()
			ve.log.Info("len webcam", len(encodedWebcamFrame))
			buffer := ve.webcamBytesBuffersPool.Get().(*bytes.Buffer)

			_, err = buffer.Write(encodedWebcamFrame)
			if err != nil {
				ve.log.Error("failed to write in webcam buffer", logger.Err(err))
				realese()
				continue
			}
			ve.userVideo.termWebcamFrameCount.Add(1)
			realese()

			select {
			case ve.webcamBufferChan <- buffer:
				closeWaitChan()
				// frameN := ve.userVideo.termWebcamFrameCount.Load()

				// if ve.onWebcamTab.Load() && (frameN%3 == 0 || frameN <= 2) {
				// 	var f string
				// 	o := ve.GetUserWebcamFrame()
				// 	ve.mu.RLock()
				// 	n := len(ve.usersVideo)
				// 	ve.mu.RUnlock()

				// 	draw.NearestNeighbor.Scale(resizedWebcamFrame, resizedWebcamFrame.Rect, webcamFrame, webcamFrame.Bounds(), draw.Over, nil)

				// 	f, err = renderLocalImg(float64(n+1), resizedWebcamFrame)
				// 	if err != nil {
				// 		ve.log.Error("failed to render img", logger.Err(err))
				// 		f = o
				// 	}
				// 	ve.userVideo.webcamFrame.Store(f)
				// }

				// ve.userVideo.mu.Lock()
				// if ve.userVideo.windowedWebcam.Load() && ve.userVideo.webcamWindow != nil {
				// 	ve.userVideo.webcamImg = webcamFrame
				// 	ve.userVideo.webcamWindow.Invalidate()
				// }
				// ve.userVideo.mu.Unlock()
				// realese()
			default:
				//realese()
				buffer.Reset()
				ve.webcamBytesBuffersPool.Put(buffer)
			}
		}

	}
}

func (ve *videoEngine) processScreen(wc chan struct{}) {
	ve.log.Info("processing screen")

	defer func() {
		if ve.waitProcessScreenChan != nil {
			ve.log.Debug("wait process screen")

			close(ve.waitProcessScreenChan)
		}
	}()

	timer := time.NewTicker(66 * time.Millisecond)
	defer timer.Stop()

	closeWaitChan := sync.OnceFunc(func() {
		close(wc)
	})

	for {
		select {
		case <-ve.stopProcessScreenChan:
			ve.log.Debug("stop process screen 1")
			return
		case <-timer.C:
			select {
			case <-ve.stopProcessScreenChan:
				ve.log.Debug("stop process screen 2")
				return
			default:
			}
			ve.mu.RLock()
			encodedRader := ve.encodedScreenReader
			ve.mu.RUnlock()
			if encodedRader == nil {
				return
			}

			encodedScreenFrame, realese, err := encodedRader.Read()
			if err != nil {
				ve.log.Error("failed to read screenFrame", logger.Err(err))
				continue
			}

		//	ve.log.Info("n", len(encodedScreenFrame))
			buffer := ve.screenBytesBuffersPool.Get().(*bytes.Buffer)

			_, err = buffer.Write(encodedScreenFrame)
			if err != nil {
				ve.log.Error("failed to write screen buffer", logger.Err(err))
				ve.screenBytesBuffersPool.Put(buffer)
				realese()
				continue
			}
			realese()
			ve.userVideo.termScreenFrameCount.Add(1)
			select {
			case ve.screenBufferChan <- buffer:
				closeWaitChan()

			default:
				buffer.Reset()
				ve.screenBytesBuffersPool.Put(buffer)
			}

		}

	}
}

func (uv *userVideo) proccessUsersWindowWebcam() {
	var ops op.Ops

	for {
		w := uv.webcamWindow
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			uv.mu.Lock()
			if w == uv.webcamWindow {
				uv.windowedWebcam.Store(false)
				uv.webcamWindow = nil
			}
			uv.mu.Unlock()
			return
		case app.FrameEvent:

			gtx := app.NewContext(&ops, e)

			uv.mu.RLock()
			im := uv.webcamImg
			uv.mu.RUnlock()
			if im != nil {

				is := im.Bounds().Size()

				if is.X > 0 && is.Y > 0 {
					scale := e.Size

					diffW := float32(scale.X) / float32(is.X)

					diffY := float32(scale.Y) / float32(is.Y)

					op.Affine(f32.Affine2D{}.Scale(f32.Pt(0, 0), f32.Pt(diffW, diffY))).Add(&ops)

					i := paint.NewImageOp(im)
					i.Add(&ops)
					paint.PaintOp{}.Add(gtx.Ops)
				}

			}
			e.Frame(gtx.Ops)
			im = nil
			ops.Reset()
		}

	}
}

func (uv *userVideo) proccessUsersWindowScreen() {
	var ops op.Ops

	for {
		w := uv.screenWindow
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			uv.mu.Lock()
			if w == uv.screenWindow {
				uv.windowedScreen.Store(false)
				uv.screenWindow = nil
			}
			uv.mu.Unlock()
			return
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			uv.mu.RLock()
			im := uv.screenImg
			uv.mu.RUnlock()
			if im != nil {

				is := im.Bounds().Size()

				if is.X > 0 && is.Y > 0 {
					scale := e.Size

					diffW := float32(scale.X) / float32(is.X)

					diffY := float32(scale.Y) / float32(is.Y)

					op.Affine(f32.Affine2D{}.Scale(f32.Pt(0, 0), f32.Pt(diffW, diffY))).Add(&ops)

					i := paint.NewImageOp(im)
					i.Add(&ops)
					paint.PaintOp{}.Add(gtx.Ops)
				}

			}
			e.Frame(gtx.Ops)
			im = nil
			ops.Reset()
		}

	}
}

func (ve *videoEngine) RenderUsersWebcam(id uuid.UUID, data []byte) {

	ve.mu.Lock()
	uv, ok := ve.usersVideo[id]
	if !ok {
		wdb := new(bytes.Buffer)
		wdb.Grow(20000)
		uv = &userVideo{
			webcamDecoderBuffer: wdb,
		}
		webcamVp8Decoder, err := vpx.NewDecoder(uv.webcamDecoderBuffer, prop.Media{})
		if err != nil {
			panic(err)
		}
		uv.webcamVp8Decoder = webcamVp8Decoder
		ve.usersVideo[id] = uv
	}
	ve.mu.Unlock()
	if uv.webcamStarted.Load() {
		ve.mu.RLock()
		n := len(ve.usersVideo)
		ve.mu.RUnlock()
		// uv.webcamDecoderBuffer.Reset()
		uv.mu.Lock()
		if uv.webcamDecoderBuffer == nil || uv.webcamVp8Decoder == nil {
			uv.mu.Unlock()
			return
		}
		_, err := uv.webcamDecoderBuffer.Write(data)
		if err != nil {
			ve.log.Error("failed to write data", logger.Err(err))
			uv.mu.Unlock()
			return
		}
		img, release, err := uv.webcamVp8Decoder.Read()
		if err != nil {
			if err.Error() == "decode failed: 5" {
				if rerr := ve.receiveWebcamKeyFrame(); rerr != nil {
					ve.log.Error("failed to receiveWebcamKeyFrame", logger.Err(rerr))
				}
			}
			//ve.log.Error("failed to decode img", logger.Err(err))
			uv.mu.Unlock()
			return
		}
		uv.mu.Unlock()
		uv.termWebcamFrameCount.Add(1)
		frameCount := uv.termWebcamFrameCount.Load()
		if ve.onWebcamTab.Load() && (frameCount%3 == 0 || frameCount <= 2) {

			if ve.userVideo.webcamStarted.Load() {
				n++
			}

			webcamFrame, err := renderTerminalImg(float64(n), img)
			if err != nil {
				ve.log.Error("failed to render img", logger.Err(err))
				uv.termWebcamFrameCount.Add(-1)
				return
			}
			if uv.webcamStarted.Load() {
				uv.webcamFrame.Store(webcamFrame)
			}

		}

		ve.mu.Lock()
		defer ve.mu.Unlock()

		uv.mu.Lock()
		if uv.windowedWebcam.Load() && uv.webcamWindow != nil {
			uv.webcamImg = img
			uv.webcamWindow.Invalidate()
		}
		release()
		uv.webcamDecoderBuffer.Reset()
		uv.mu.Unlock()
	}

}

func (ve *videoEngine) RenderUsersScreen(id uuid.UUID, data []byte) {
	ve.mu.Lock()
	uv, ok := ve.usersVideo[id]
	if !ok {
		sdb := new(bytes.Buffer)
		sdb.Grow(150000)
		uv = &userVideo{
			screenDecoderBuffer: sdb,
		}
		screenVp8Decoder, err := vpx.NewDecoder(uv.screenDecoderBuffer, prop.Media{})
		if err != nil {
			panic(err)
		}
		uv.screenVp8Decoder = screenVp8Decoder
		ve.usersVideo[id] = uv
	}
	ve.mu.Unlock()

	if uv.screenStarted.Load() {
		ve.mu.RLock()
		n := len(ve.usersVideo)
		ve.mu.RUnlock()

		uv.mu.Lock()
		if uv.screenDecoderBuffer == nil || uv.screenVp8Decoder == nil {
			uv.mu.Unlock()
			return
		}
		_, err := uv.screenDecoderBuffer.Write(data)
		if err != nil {
			ve.log.Error("failed to write data", logger.Err(err))
			uv.mu.Unlock()
			return
		}
		img, release, err := uv.screenVp8Decoder.Read()
		if err != nil {
			if err.Error() == "decode failed: 5" {
				if rerr := ve.receiveScreenKeyFrame(); rerr != nil {
					ve.log.Error("failed to receiveScreenKeyFrame", logger.Err(rerr))
				}
			}
			//ve.log.Error("failed to decode img", logger.Err(err))
			uv.mu.Unlock()
			return
		}
		uv.mu.Unlock()
		uv.termScreenFrameCount.Add(1)
		frameCount := uv.termScreenFrameCount.Load()
		if ve.onScreenTab.Load() && (frameCount%3 == 0 || frameCount <= 2) {

			if ve.userVideo.screenStarted.Load() {
				n++
			}
			screenFrame, err := renderTerminalImg(float64(n), img)
			if err != nil {
				ve.log.Error("failed to render img", logger.Err(err))
				uv.termScreenFrameCount.Add(-1)
				return
			}
			if uv.screenStarted.Load() {
				uv.screenFrame.Store(screenFrame)
			}
		}

		ve.mu.Lock()
		defer ve.mu.Unlock()

		uv.mu.Lock()
		if uv.windowedScreen.Load() && uv.screenWindow != nil {
			uv.screenImg = img
			uv.screenWindow.Invalidate()
		}
		release()
		uv.screenDecoderBuffer.Reset()
		uv.mu.Unlock()

	}

}

func (ve *videoEngine) OnOffUsersWebcam(res bool, id uuid.UUID) {
	ve.mu.Lock()
	defer ve.mu.Unlock()
	uv, ok := ve.usersVideo[id]
	if !ok {
		uv = &userVideo{}
		ve.usersVideo[id] = uv
	}
	uv.mu.Lock()
	if !res {

		if uv.windowedWebcam.Load() && uv.webcamWindow != nil {
			uv.webcamWindow.Perform(system.ActionClose)
		}
		uv.webcamImg = nil
		uv.webcamFrame.Store("")
		if err := uv.webcamVp8Decoder.Close(); err != nil {
			ve.log.Error("failed to close webcam user decoder", logger.Err(err))
		}
		uv.webcamDecoderBuffer = nil
		uv.webcamVp8Decoder = nil

	} else {
		wdb := new(bytes.Buffer)
		wdb.Grow(20000)
		uv.webcamDecoderBuffer = wdb
		webcamVp8Decoder, err := vpx.NewDecoder(uv.webcamDecoderBuffer, prop.Media{})
		if err != nil {
			panic(err)
		}
		uv.webcamVp8Decoder = webcamVp8Decoder
	}
	uv.mu.Unlock()
	uv.webcamStarted.Store(res)

}

func (ve *videoEngine) OnOffUsersScreen(res bool, id uuid.UUID) {
	ve.mu.Lock()
	defer ve.mu.Unlock()
	uv, ok := ve.usersVideo[id]
	if !ok {
		uv = &userVideo{}
		ve.usersVideo[id] = uv
	}

	uv.mu.Lock()
	if !res {
		if uv.windowedScreen.Load() && uv.screenWindow != nil {
			uv.screenWindow.Perform(system.ActionClose)
		}
		uv.screenImg = nil
		uv.screenFrame.Store("")
		if err := uv.screenVp8Decoder.Close(); err != nil {
			ve.log.Error("failed to close screen user decoder", logger.Err(err))
		}
		uv.screenDecoderBuffer = nil
		uv.screenVp8Decoder = nil

	} else {
		sdb := new(bytes.Buffer)
		sdb.Grow(150000)
		uv.screenDecoderBuffer = sdb
		screenVp8Decoder, err := vpx.NewDecoder(uv.screenDecoderBuffer, prop.Media{})
		if err != nil {
			panic(err)
		}
		uv.screenVp8Decoder = screenVp8Decoder
	}
	uv.mu.Unlock()

	uv.screenStarted.Store(res)
}

func (ve *videoEngine) GetUsersWebcamFramesTerminal() map[uuid.UUID]string {
	ve.mu.RLock()
	defer ve.mu.RUnlock()
	var (
		f  string
		ok bool
	)
	frames := make(map[uuid.UUID]string, len(ve.usersVideo))
	for i, uv := range ve.usersVideo {
		f, ok = uv.webcamFrame.Load().(string)
		if ok && f != "" {
			frames[i] = f
		}
	}

	return frames
}

func (ve *videoEngine) SendWebcamKeyFrame() error {
	ve.log.Debug("sending webcam key frame")
	ve.mu.Lock()
	defer ve.mu.Unlock()
	if ve.webcamKeyFrameCtrl != nil {
		if err := ve.webcamKeyFrameCtrl.ForceKeyFrame(); err != nil {
			ve.log.Error("failed to force webcam key frame", logger.Err(err))
			return err
		}
	}
	return nil
}

func (ve *videoEngine) SendScreenKeyFrame() error {
	ve.log.Debug("sending screen key frame")
	ve.mu.Lock()
	defer ve.mu.Unlock()
	if ve.screenKeyFrameCtrl != nil {
		if err := ve.screenKeyFrameCtrl.ForceKeyFrame(); err != nil {
			ve.log.Error("failed to force screen key frame", logger.Err(err))
			return err
		}
	}
	return nil
}

func (ve *videoEngine) GetUsersScreenFramesTerminal() map[uuid.UUID]string {
	ve.mu.RLock()
	defer ve.mu.RUnlock()
	var (
		f  string
		ok bool
	)
	frames := make(map[uuid.UUID]string, len(ve.usersVideo))
	for i, uv := range ve.usersVideo {

		f, ok = uv.screenFrame.Load().(string)
		if ok && f != "" {
			frames[i] = f
		}
	}

	return frames
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

	termW, termH, err := term.GetSize(int(os.Stdout.Fd()))
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

	termW, termH, err := term.GetSize(int(os.Stdout.Fd()))
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

func (ve *videoEngine) sendWebcam() {
	ve.log.Info("sending webcam")

	for {
		select {
		case <-ve.stopSendVideoChan:
			ve.log.Info("webcam sending stopped")
			return
		case webcamBuffer := <-ve.webcamBufferChan:
			data := webcamBuffer.Bytes()
			if len(data) == 0 {
				ve.log.Error("zero webcam data")
				continue
			}
			ve.mu.RLock()
			netw := ve.netw
			ve.mu.RUnlock()
			if ve.connected.Load() && ve.userVideo.webcamStarted.Load() && netw != nil {
				if err := netw.SendWebcamData(data); err != nil {
					ve.log.Error("failed to send webcam data", logger.Err(err))
				}
			}
			webcamBuffer.Reset()
			ve.webcamBytesBuffersPool.Put(webcamBuffer)
		}
	}
}

func (ve *videoEngine) sendScreen() {
	ve.log.Info("sending screen")

	for {
		select {
		case <-ve.stopSendVideoChan:
			ve.log.Info("screen sending stopped")
			return
		case screenBuffer := <-ve.screenBufferChan:
			data := screenBuffer.Bytes()
			if len(data) == 0 {
				ve.log.Error("zero screen data")
				data = nil
				continue
			}
			ve.mu.RLock()
			netw := ve.netw
			ve.mu.RUnlock()
			if ve.connected.Load() && ve.userVideo.screenStarted.Load() && netw != nil {
				if err := netw.SendScreenData(data); err != nil {
					ve.log.Error("failed to send screen data", logger.Err(err))
				}
			}
			screenBuffer.Reset()
			data = nil

			ve.screenBytesBuffersPool.Put(screenBuffer)
		}
	}
}

func (ve *videoEngine) receiveWebcamKeyFrame() error {
	ve.mu.RLock()
	netw := ve.netw
	ve.mu.RUnlock()
	if ve.connected.Load() && netw != nil {
		wkfe, err := networking.WebcamKeyFrameEvent()
		if err != nil {
			ve.log.Error("failed to create webcam key frame event", logger.Err(err))
			return err
		}
		if err := netw.NewEvent(wkfe); err != nil {
			ve.log.Error("failed to send webcam key frame event", logger.Err(err))
			return err
		}
	}
	return nil
}

func (ve *videoEngine) receiveScreenKeyFrame() error {
	ve.mu.RLock()
	netw := ve.netw
	ve.mu.RUnlock()
	if ve.connected.Load() && netw != nil {
		skfe, err := networking.ScreenKeyFrameEvent()
		if err != nil {
			ve.log.Error("failed to create screen key frame event", logger.Err(err))
			return err
		}
		if err := netw.NewEvent(skfe); err != nil {
			ve.log.Error("failed to send screen key frame event", logger.Err(err))
			return err
		}
	}
	return nil
}

func (ve *videoEngine) Stop() {

	close(ve.stopSendVideoChan)

	if ve.webcamTrack != nil {
		if err := ve.webcamTrack.Close(); err != nil {
			ve.log.Error("failed to close webcamTrack", logger.Err(err))
		}
		ve.webcamTrack = nil
	}

	if ve.encodedWebcamReader != nil {
		if err := ve.encodedWebcamReader.Close(); err != nil {
			ve.log.Error("failed to close encodedWebcamReader", logger.Err(err))
		}
		ve.encodedWebcamReader = nil
	}

	//ve.webcamReader = nil

	if ve.screen != nil {
		if err := ve.screen.Close(); err != nil {
			ve.log.Error("failed to close screen", logger.Err(err))
		}
		ve.screen = nil
	}

	if ve.encodedScreenReader != nil {
		if err := ve.encodedScreenReader.Close(); err != nil {
			ve.log.Error("failed to close screenTrack", logger.Err(err))
		}

		ve.encodedScreenReader = nil
	}

}

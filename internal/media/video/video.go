package video

import (
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
	"github.com/kolesa-team/go-webp/decoder"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"github.com/pion/mediadevices"
	_ "github.com/pion/mediadevices/pkg/driver/camera"
	_ "github.com/pion/mediadevices/pkg/driver/screen"
	"github.com/pion/mediadevices/pkg/io/video"
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
	OnOffUsersScreen(res bool, id uuid.UUID)
	OnOffUsersWebcam(res bool, id uuid.UUID)
	OnOffWebcamWindow() (bool, error)
	OnOffScreenWindow() (bool, error)
	SwitchOnWebcamTab(res bool)
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
	webcamReader video.Reader
	webcamBuffer *bytes.Buffer
	webcamTrack  *mediadevices.VideoTrack

	screenReader video.Reader
	screenBuffer *bytes.Buffer
	screenTrack  *mediadevices.VideoTrack

	netw networking.Networking

	webcamFrameChan chan []byte
	screenFrameChan chan []byte

	stopSendVideoChan     chan struct{}
	stopProcessWebcamChan chan struct{}
	waitProcessWebcamChan chan struct{}
	stopProcessScreenChan chan struct{}
	waitProcessScreenChan chan struct{}

	webcamBytesBuffersPool sync.Pool
	screenBytesBuffersPool sync.Pool

	connected atomic.Bool

	onWebcamTab atomic.Bool
	onScreenTab atomic.Bool

	usersVideo map[uuid.UUID]*userVideo

	userVideo *userVideo

	mu sync.RWMutex

	log *logger.Logger
}

type VideoSetup struct {
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
}

func NewVideoEngine(l *logger.Logger, vs VideoSetup) (VideoEngine, error) {
	log := l.AddOp("videoEngine")

	ve := &videoEngine{
		log:               log,
		webcamBuffer:      bytes.NewBuffer(make([]byte, 0, 40000)),
		screenBuffer:      bytes.NewBuffer(make([]byte, 0, 150000)),
		webcamFrameChan:   make(chan []byte, 1),
		screenFrameChan:   make(chan []byte, 1),
		stopSendVideoChan: make(chan struct{}, 1),
		usersVideo:        make(map[uuid.UUID]*userVideo, 3),
		userVideo:         &userVideo{},
		webcamBytesBuffersPool: sync.Pool{
			New: func() any {
				buf := make([]byte, 40000)
				return buf
			},
		},
		screenBytesBuffersPool: sync.Pool{
			New: func() any {
				buf := make([]byte, 150000)
				return buf
			},
		},
	}

	go ve.sendWebcam()
	go ve.sendScreen()

	return ve, nil
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
	ve.userVideo.mu.Unlock()
	ve.userVideo.webcamImg = nil
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
	ve.userVideo.mu.Unlock()
	ve.userVideo.screenImg = nil
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
	uv.mu.Unlock()
	uv.screenImg = nil
	return false, nil

}

func (ve *videoEngine) OnOffWebcam() (bool, error) {
	s := ve.userVideo.webcamStarted.Load()

	if !s == true {

		stream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
			Video: func(mtc *mediadevices.MediaTrackConstraints) {},
		})
		if err != nil {
			ve.log.Error("failed to get stream webcam", logger.Err(err))
			return false, err
		}

		track := stream.GetVideoTracks()[0]

		videoTrack := track.(*mediadevices.VideoTrack)

		videoReader := videoTrack.NewReader(false)

		ve.webcamReader = videoReader
		ve.webcamTrack = videoTrack

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

		ve.webcamReader = nil
		ve.mu.Unlock()
	}

	ve.userVideo.webcamFrame.Store("")
	ve.userVideo.webcamStarted.Store(!s)
	return !s, nil
}

func (ve *videoEngine) OnOffScreen() (bool, error) {
	s := ve.userVideo.screenStarted.Load()

	if !s == true {

		stream, err := mediadevices.GetDisplayMedia(mediadevices.MediaStreamConstraints{
			Video: func(mtc *mediadevices.MediaTrackConstraints) {},
		})
		if err != nil {
			ve.log.Error("failed to get stream screen", logger.Err(err))
			return false, err
		}

		track := stream.GetVideoTracks()[0]

		videoTrack := track.(*mediadevices.VideoTrack)

		videoReader := videoTrack.NewReader(false)

		ve.screenReader = videoReader
		ve.screenTrack = videoTrack

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
		ve.waitProcessScreenChan = nil
		ve.stopProcessScreenChan = nil
		if ve.screenTrack != nil {
			ve.log.Info("closing screen track")
			if err := ve.screenTrack.Close(); err != nil {
				ve.log.Error("failed to close screen track", logger.Err(err))
			}
			ve.screenTrack = nil
		}

		ve.screenReader = nil
		ve.mu.Unlock()

	}

	ve.userVideo.screenFrame.Store("")
	ve.userVideo.screenStarted.Store(!s)
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
	ve.webcamReader = nil
	ve.stopProcessScreenChan = nil
	if ve.screenTrack != nil {
		ve.log.Info("close screen track when disc")
		if err := ve.screenTrack.Close(); err != nil {
			ve.log.Error("failed to close screen track", logger.Err(err))
		}
		ve.screenTrack = nil
	}
	ve.screenReader = nil
}

func (ve *videoEngine) processWebcam(wc chan struct{}) {
	ve.log.Info("processing webcam")

	defer func() {
		if ve.waitProcessWebcamChan != nil {
			close(ve.waitProcessWebcamChan)
		}
	}()

	timer := time.NewTicker(66 * time.Millisecond)

	closeWaitChan := sync.OnceFunc(func() {
		close(wc)
	})

	eOpt, err := encoder.NewLossyEncoderOptions(encoder.PresetPicture, 75)
	if err != nil {
		panic(err)
	}

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
			reader := ve.webcamReader
			ve.mu.RUnlock()
			if reader == nil {
				return
			}
			webcamFrame, realese, err := reader.Read()
			if err != nil {
				ve.log.Error("failed to read webcamFrame", logger.Err(err))
				continue
			}

			resizedWebcamFrame := image.NewNRGBA(image.Rect(0, 0, 360, 160))

			draw.NearestNeighbor.Scale(resizedWebcamFrame, resizedWebcamFrame.Rect, webcamFrame, webcamFrame.Bounds(), draw.Over, nil)
			realese()
			ve.webcamBuffer.Reset()

			if err := webp.Encode(ve.webcamBuffer, resizedWebcamFrame, eOpt); err != nil {
				ve.log.Error("failed to encode webcamFrame", logger.Err(err))
				return
			}

			buf := ve.webcamBuffer.Bytes()

			buffer := ve.webcamBytesBuffersPool.Get().([]byte)
			copy(buffer, buf)

			n := len(buf)
			select {
			case ve.webcamFrameChan <- buffer[:n]:
				if ve.onWebcamTab.Load() {
					var f string
					o := ve.GetUserWebcamFrame()
					ve.mu.RLock()
					n := len(ve.usersVideo)
					ve.mu.RUnlock()
					f, err = renderLocalImg(float64(n+1), resizedWebcamFrame)
					if err != nil {
						ve.log.Error("failed to render img", logger.Err(err))
						f = o
					}
					ve.userVideo.webcamFrame.Store(f)
				}

				closeWaitChan()

				ve.userVideo.mu.Lock()
				if ve.userVideo.windowedWebcam.Load() && ve.userVideo.webcamWindow != nil {
					ve.userVideo.webcamImg = webcamFrame
					ve.userVideo.webcamWindow.Invalidate()
				}
				ve.userVideo.mu.Unlock()
			default:
				ve.webcamBytesBuffersPool.Put(buffer[:40000])
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

	closeWaitChan := sync.OnceFunc(func() {
		close(wc)
	})

	eOpt, err := encoder.NewLossyEncoderOptions(encoder.PresetPicture, 75)
	if err != nil {
		panic(err)
	}

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
			reader := ve.screenReader
			ve.mu.RUnlock()
			if reader == nil {
				return
			}
			screenFrame, realese, err := reader.Read()
			if err != nil {
				ve.log.Error("failed to read screenFrame", logger.Err(err))
				continue
			}

			resizedScreenFrame := image.NewNRGBA(image.Rect(0, 0, 360, 160))

			draw.NearestNeighbor.Scale(resizedScreenFrame, resizedScreenFrame.Rect, screenFrame, screenFrame.Bounds(), draw.Over, nil)

			ve.screenBuffer.Reset()

			if err := webp.Encode(ve.screenBuffer, screenFrame, eOpt); err != nil {
				ve.log.Error("failed to encode screenFrame", logger.Err(err))
				return
			}

			realese()

			buf := ve.screenBuffer.Bytes()

			n := len(buf)

			buffer := ve.screenBytesBuffersPool.Get().([]byte)
			copy(buffer, buf)

			select {
			case ve.screenFrameChan <- buffer[:n]:
				if ve.onScreenTab.Load() {
					var f string
					o := ve.GetUserScreenFrame()
					ve.mu.RLock()
					n := len(ve.usersVideo)
					ve.mu.RUnlock()
					f, err = renderLocalImg(float64(n+1), resizedScreenFrame)
					if err != nil {
						ve.log.Error("failed to render img", logger.Err(err))
						f = o
					}
					ve.userVideo.screenFrame.Store(f)
				}

				closeWaitChan()

				ve.userVideo.mu.Lock()
				if ve.userVideo.windowedScreen.Load() && ve.userVideo.screenWindow != nil {
					ve.userVideo.screenImg = screenFrame
					ve.userVideo.screenWindow.Invalidate()
				}
				ve.userVideo.mu.Unlock()
			default:
				ve.screenBytesBuffersPool.Put(buffer[:150000])
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
		}

	}
}

func (ve *videoEngine) RenderUsersWebcam(id uuid.UUID, data []byte) {

	ve.mu.Lock()
	uv, ok := ve.usersVideo[id]
	if !ok {
		uv = &userVideo{}
		ve.usersVideo[id] = uv
	}
	ve.mu.Unlock()
	if uv.webcamStarted.Load() {
		ve.mu.RLock()
		n := len(ve.usersVideo)
		ve.mu.RUnlock()
		img, err := webp.Decode(bytes.NewReader(data), &decoder.Options{})
		if err != nil {
			ve.log.Error("failed to decode img", logger.Err(err))
			return
		}

		if ve.onWebcamTab.Load() {

			if ve.userVideo.webcamStarted.Load() {
				n++
			}

			webcamFrame, err := renderTerminalImg(float64(n), img)
			if err != nil {
				ve.log.Error("failed to render img", logger.Err(err))
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
		uv.mu.Unlock()
	}

}

func (ve *videoEngine) RenderUsersScreen(id uuid.UUID, data []byte) {
	ve.mu.Lock()
	uv, ok := ve.usersVideo[id]
	if !ok {
		uv = &userVideo{}
		ve.usersVideo[id] = uv
	}
	ve.mu.Unlock()
	if uv.screenStarted.Load() {
		ve.mu.RLock()
		n := len(ve.usersVideo)
		ve.mu.RUnlock()

		img, err := webp.Decode(bytes.NewReader(data), &decoder.Options{})
		if err != nil {
			ve.log.Error("failed to decode img", logger.Err(err))
			return
		}

		if ve.onScreenTab.Load() {

			if ve.userVideo.screenStarted.Load() {
				n++
			}
			screenFrame, err := renderTerminalImg(float64(n), img)
			if err != nil {
				ve.log.Error("failed to render img", logger.Err(err))
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

	uv.webcamStarted.Store(res)
	if !res {
		uv.mu.Lock()
		if uv.windowedWebcam.Load() && uv.webcamWindow != nil {
			uv.webcamWindow.Perform(system.ActionClose)
		}
		uv.webcamImg = nil
		uv.webcamFrame.Store("")
		uv.mu.Unlock()
	}

}

func (ve *videoEngine) OnOffUsersScreen(res bool, id uuid.UUID) {
	ve.mu.Lock()
	defer ve.mu.Unlock()
	uv, ok := ve.usersVideo[id]
	if !ok {
		uv = &userVideo{}
		ve.usersVideo[id] = uv
	}

	uv.screenStarted.Store(res)
	if !res {

		uv.mu.Lock()
		if uv.windowedScreen.Load() && uv.screenWindow != nil {
			uv.screenWindow.Perform(system.ActionClose)
		}
		uv.screenImg = nil
		uv.screenFrame.Store("")
		uv.mu.Unlock()
	}

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
		case webcamFrame := <-ve.webcamFrameChan:
			if ve.connected.Load() && ve.userVideo.webcamStarted.Load() && ve.netw != nil {
				if err := ve.netw.SendWebcamData(webcamFrame); err != nil {
					ve.log.Error("failed to send webcam data", logger.Err(err))
				}
			}
			ve.webcamBytesBuffersPool.Put(webcamFrame[:40000])
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
		case screenFrame := <-ve.screenFrameChan:
			if ve.connected.Load() && ve.userVideo.screenStarted.Load() && ve.netw != nil {
				if err := ve.netw.SendScreenData(screenFrame); err != nil {
					ve.log.Error("failed to send screen data", logger.Err(err))
				}
			}
			ve.screenBytesBuffersPool.Put(screenFrame[:150000])
		}
	}
}

func (ve *videoEngine) Stop() {

	close(ve.stopSendVideoChan)

	if ve.webcamTrack != nil {
		if err := ve.webcamTrack.Close(); err != nil {
			ve.log.Error("failed to close webcamTrack", logger.Err(err))
		}
	}

	ve.webcamReader = nil

	if ve.screenTrack != nil {
		if err := ve.screenTrack.Close(); err != nil {
			ve.log.Error("failed to close screenTrack", logger.Err(err))
		}
	}

	ve.screenReader = nil

}

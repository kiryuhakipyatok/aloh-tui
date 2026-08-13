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
	"github.com/pion/mediadevices/pkg/prop"
)

type VideoEngine interface {
	SetNetworking(netw networking.Networking)
	SetConnected()
	SetDisconnected()
	RenderUsersWebcam(id uuid.UUID, data []byte)
	OnOffUsersWebcamWindow(id uuid.UUID, nickname string) (bool, error)
	OnOffUsersScreenWindow(id uuid.UUID, nickname string) (bool, error)
	GetUsersWebcamFramesTerminal() map[uuid.UUID]string
	GetUsersScreenFramesTerminal() map[uuid.UUID]string
	RemoveUserFromUsersVideo(id uuid.UUID)
	OnOffWebcamWindow() (bool, error)
	OnOffScreenWindow() (bool, error)
	OnOffWebcam() (bool, error)
	OnOffScreen() (bool, error)
	GetUserWebcamFrame() string
	GetUserScreenFrame() string
	IsStarted() bool
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
	stopProcessScreenChan chan struct{}

	bytesBuffersPool sync.Pool

	connected atomic.Bool
	started   atomic.Bool

	usersVideo map[uuid.UUID]*userVideo

	userVideo *userVideo

	mu sync.RWMutex

	log *logger.Logger
}

type VideoSetup struct {
}

type userVideo struct {
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
		log:                   log,
		webcamBuffer:          bytes.NewBuffer(make([]byte, 0, 2000)),
		screenBuffer:          bytes.NewBuffer(make([]byte, 0, 2000)),
		webcamFrameChan:       make(chan []byte, 1),
		screenFrameChan:       make(chan []byte, 1),
		stopSendVideoChan:     make(chan struct{}, 1),
		stopProcessWebcamChan: make(chan struct{}, 1),
		usersVideo:            make(map[uuid.UUID]*userVideo, 3),
		userVideo:             &userVideo{},
		bytesBuffersPool: sync.Pool{
			New: func() any {
				buf := make([]byte, 1350)
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

func (ve *videoEngine) IsStarted() bool {
	return ve.started.Load()
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
	s := ve.started.Load()

	if !s == true {

		stream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
			Video: func(mtc *mediadevices.MediaTrackConstraints) {
				mtc.Width = prop.Int(360)
				mtc.Height = prop.Int(160)
			},
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
		waitChan := make(chan struct{}, 1)
		go ve.processWebcam(waitChan)
		<-waitChan

	} else {
		ve.userVideo.mu.Lock()
		if ve.userVideo.windowedWebcam.Load() && ve.userVideo.webcamWindow != nil {
			ve.userVideo.webcamWindow.Perform(system.ActionClose)
		}
		ve.userVideo.mu.Unlock()
		close(ve.stopProcessWebcamChan)
		ve.stopProcessWebcamChan = nil
		if ve.webcamTrack != nil {
			if err := ve.webcamTrack.Close(); err != nil {
				ve.log.Error("failed to close webcam track", logger.Err(err))
			}
		}

		ve.webcamReader = nil

	}

	ve.userVideo.webcamFrame.Store("")
	ve.started.Store(!s)
	return !s, nil
}

func (ve *videoEngine) OnOffScreen() (bool, error) {
	s := ve.started.Load()

	if !s == true {

		stream, err := mediadevices.GetDisplayMedia(mediadevices.MediaStreamConstraints{
			Video: func(mtc *mediadevices.MediaTrackConstraints) {
				mtc.Width = prop.Int(360)
				mtc.Height = prop.Int(160)
			},
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
		waitChan := make(chan struct{}, 1)
		go ve.processScreen(waitChan)
		<-waitChan

	} else {
		ve.userVideo.mu.Lock()
		if ve.userVideo.windowedWebcam.Load() && ve.userVideo.webcamWindow != nil {
			ve.userVideo.webcamWindow.Perform(system.ActionClose)
		}
		ve.userVideo.mu.Unlock()
		close(ve.stopProcessWebcamChan)
		ve.stopProcessWebcamChan = nil
		if ve.webcamTrack != nil {
			if err := ve.webcamTrack.Close(); err != nil {
				ve.log.Error("failed to close webcam track", logger.Err(err))
			}
		}

		ve.webcamReader = nil

	}

	ve.userVideo.webcamFrame.Store("")
	ve.started.Store(!s)
	return !s, nil
}

func (ve *videoEngine) SetDisconnected() {
	ve.connected.Store(false)
	ve.started.Store(false)
	ve.userVideo.webcamFrame.Store("")
	ve.mu.Lock()

	ve.userVideo.mu.Lock()
	if ve.userVideo.windowedWebcam.Load() && ve.userVideo.webcamWindow != nil {
		ve.userVideo.webcamWindow.Perform(system.ActionClose)
	}
	ve.userVideo.mu.Unlock()

	for _, uv := range ve.usersVideo {
		uv.mu.Lock()
		if uv.webcamWindow != nil {
			uv.webcamWindow.Perform(system.ActionClose)
		}
		uv.mu.Unlock()
	}

	clear(ve.usersVideo)

	if ve.stopProcessWebcamChan != nil {
		close(ve.stopProcessWebcamChan)
		ve.stopProcessWebcamChan = nil
	}

	if ve.webcamTrack != nil {
		if err := ve.webcamTrack.Close(); err != nil {
			ve.log.Error("failed to close webcam track", logger.Err(err))
		}
	}

	ve.webcamReader = nil
	ve.mu.Unlock()
}

func (ve *videoEngine) processWebcam(wc chan struct{}) {
	ve.log.Info("processing webcam")
	timer := time.NewTicker(50 * time.Millisecond)

	closeWaitChan := sync.OnceFunc(func() {
		close(wc)
	})

	eOpt, err := encoder.NewLossyEncoderOptions(encoder.PresetPicture, 1)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-ve.stopProcessWebcamChan:
			return
		case <-timer.C:
			if ve.webcamReader == nil {
				return
			}
			webcamFrame, realese, err := ve.webcamReader.Read()
			if err != nil {
				ve.log.Error("failed to read webcamFrame", logger.Err(err))
				continue
			}

			ve.webcamBuffer.Reset()

			if err := webp.Encode(ve.webcamBuffer, webcamFrame, eOpt); err != nil {
				ve.log.Error("failed to encode webcamFrame", logger.Err(err))
				continue
			}

			realese()

			buf := ve.webcamBuffer.Bytes()

			buffer := ve.bytesBuffersPool.Get().([]byte)
			copy(buffer, buf)

			n := len(buf)

			select {
			case ve.webcamFrameChan <- buffer[:n]:
				var f string
				o := ve.GetUserWebcamFrame()
				ve.mu.RLock()
				n := len(ve.usersVideo)
				ve.mu.RUnlock()
				f, err = renderLocalImg(n, webcamFrame)
				if err != nil {
					ve.log.Error("failed to render img", logger.Err(err))
					f = o
				}

				closeWaitChan()
				ve.userVideo.webcamFrame.Store(f)
				ve.userVideo.mu.Lock()
				if ve.userVideo.windowedWebcam.Load() && ve.userVideo.webcamWindow != nil {
					ve.userVideo.webcamImg = webcamFrame
					ve.userVideo.webcamWindow.Invalidate()
				}
				ve.userVideo.mu.Unlock()
			default:
				ve.bytesBuffersPool.Put(buffer[:1350])
			}
		}

	}
}

func (ve *videoEngine) processScreen(wc chan struct{}) {
	ve.log.Info("processing screen")
	timer := time.NewTicker(50 * time.Millisecond)

	closeWaitChan := sync.OnceFunc(func() {
		close(wc)
	})

	eOpt, err := encoder.NewLossyEncoderOptions(encoder.PresetPicture, 1)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-ve.stopProcessScreenChan:
			return
		case <-timer.C:
			if ve.screenReader == nil {
				return
			}
			screenFrame, realese, err := ve.screenReader.Read()
			if err != nil {
				ve.log.Error("failed to read screenFrame", logger.Err(err))
				continue
			}

			ve.screenBuffer.Reset()

			if err := webp.Encode(ve.screenBuffer, screenFrame, eOpt); err != nil {
				ve.log.Error("failed to encode screenFrame", logger.Err(err))
				continue
			}

			realese()

			buf := ve.screenBuffer.Bytes()

			buffer := ve.bytesBuffersPool.Get().([]byte)
			copy(buffer, buf)

			n := len(buf)

			select {
			case ve.screenFrameChan <- buffer[:n]:
				var f string
				o := ve.GetUserScreenFrame()
				ve.mu.RLock()
				n := len(ve.usersVideo)
				ve.mu.RUnlock()
				f, err = renderLocalImg(n, screenFrame)
				if err != nil {
					ve.log.Error("failed to render img", logger.Err(err))
					f = o
				}

				closeWaitChan()
				ve.userVideo.screenFrame.Store(f)
				ve.userVideo.mu.Lock()
				if ve.userVideo.windowedScreen.Load() && ve.userVideo.screenWindow != nil {
					ve.userVideo.screenImg = screenFrame
					ve.userVideo.screenWindow.Invalidate()
				}
				ve.userVideo.mu.Unlock()
			default:
				ve.bytesBuffersPool.Put(buffer[:1350])
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
	n := len(ve.usersVideo)
	ve.mu.Unlock()
	img, err := webp.Decode(bytes.NewReader(data), &decoder.Options{})
	if err != nil {
		ve.log.Error("failed to decode img", logger.Err(err))
		return
	}
	webcamFrame, err := renderTerminalImg(n, img)
	if err != nil {
		ve.log.Error("failed to render img", logger.Err(err))
		return
	}
	ve.mu.Lock()
	defer ve.mu.Unlock()
	uv.webcamFrame.Store(webcamFrame)

	uv.mu.Lock()
	if uv.windowedWebcam.Load() && uv.webcamWindow != nil {
		uv.webcamImg = img
		uv.webcamWindow.Invalidate()
	}
	uv.mu.Unlock()

}

func (ve *videoEngine) RemoveUserFromUsersVideo(id uuid.UUID) {
	ve.mu.Lock()
	defer ve.mu.Unlock()
	uv, ok := ve.usersVideo[id]
	if ok {
		uv.mu.Lock()
		if uv.windowedWebcam.Load() && uv.webcamWindow != nil {
			uv.webcamWindow.Perform(system.ActionClose)
		}
		uv.mu.Unlock()
		delete(ve.usersVideo, id)
	}
}

func (ve *videoEngine) GetUsersWebcamFramesTerminal() map[uuid.UUID]string {
	ve.mu.RLock()
	defer ve.mu.RUnlock()
	frames := make(map[uuid.UUID]string, len(ve.usersVideo))
	for i, uv := range ve.usersVideo {
		frames[i] = uv.webcamFrame.Load().(string)
	}

	return frames
}

func (ve *videoEngine) GetUsersScreenFramesTerminal() map[uuid.UUID]string {
	ve.mu.RLock()
	defer ve.mu.RUnlock()
	frames := make(map[uuid.UUID]string, len(ve.usersVideo))
	for i, uv := range ve.usersVideo {
		frames[i] = uv.screenFrame.Load().(string)
	}

	return frames
}

func renderTerminalImg(n int, img image.Image) (string, error) {
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

	imageWidget.SetSizeWithCorrection(int(float64(termW)/(float64(n)*1.15)), int(float64(termH)/(float64(n))))
	textMsg, err := imageWidget.Render()
	if err != nil {
		return "", err
	}

	return textMsg, nil
}

func renderLocalImg(n int, img image.Image) (string, error) {
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

	imageWidget.SetSizeWithCorrection(int(float64(termW)/(float64(n)*1.15)), int(float64(termH)/(float64(n))))

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
			if ve.connected.Load() && ve.started.Load() && ve.netw != nil {
				if err := ve.netw.SendVideoData(webcamFrame); err != nil {
					ve.log.Error("failed to send webcam data", logger.Err(err))
				}
			}
			ve.bytesBuffersPool.Put(webcamFrame[:1350])
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
			if ve.connected.Load() && ve.started.Load() && ve.netw != nil {
				if err := ve.netw.SendVideoData(screenFrame); err != nil {
					ve.log.Error("failed to send screen data", logger.Err(err))
				}
			}
			ve.bytesBuffersPool.Put(screenFrame[:1350])
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

}

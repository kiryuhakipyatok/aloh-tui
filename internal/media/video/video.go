package video

import (
	"aloh-tui/internal/networking"
	"aloh-tui/pkg/errs"
	"aloh-tui/pkg/logger"
	"bytes"
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
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

type VideoEngine interface {
	SetNetworking(netw networking.Networking)
	SetConnected()
	SetDisconnected()
	RenderUsersWebcam(id uuid.UUID, data []byte)
	OnOffUsersWindow(id uuid.UUID, nickname string) (bool, error)
	GetUsersFramesTerminal() map[uuid.UUID]string
	RemoveUserFromUsersVideo(id uuid.UUID)
	OnOffWebcam() (bool, error)
	GetUserFrame() string
	IsStarted() bool
	Stop()
}

type videoEngine struct {
	webcamReader video.Reader
	webcamBuffer *bytes.Buffer
	webcamTrack  *mediadevices.VideoTrack

	netw networking.Networking

	videoFrameChan chan []byte

	stopSendVideoChan     chan struct{}
	stopProcessWebcamChan chan struct{}

	bytesBuffersPool sync.Pool

	connected atomic.Bool
	started   atomic.Bool

	usersVideo map[uuid.UUID]*userVideo

	userFrame atomic.Value

	mu sync.RWMutex

	log *logger.Logger
}

type VideoSetup struct {
}

type userVideo struct {
	frame    string
	img      image.Image
	mu       sync.RWMutex
	window   *app.Window
	windowed atomic.Bool
}

func NewVideoEngine(l *logger.Logger, vs VideoSetup) (VideoEngine, error) {
	log := l.AddOp("videoEngine")

	ve := &videoEngine{
		log:                   log,
		webcamBuffer:          bytes.NewBuffer(make([]byte, 0, 2000)),
		videoFrameChan:        make(chan []byte, 1),
		stopSendVideoChan:     make(chan struct{}, 1),
		stopProcessWebcamChan: make(chan struct{}, 1),
		usersVideo:            make(map[uuid.UUID]*userVideo, 3),
		bytesBuffersPool: sync.Pool{
			New: func() any {
				buf := make([]byte, 1350)
				return buf
			},
		},
	}

	go ve.sendVideo()

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

func (ve *videoEngine) GetUserFrame() string {
	f, ok := ve.userFrame.Load().(string)
	if !ok {
		return ""
	}
	return f
}

func (ve *videoEngine) OnOffUsersWindow(id uuid.UUID, nickname string) (bool, error) {
	ve.mu.Lock()
	defer ve.mu.Unlock()
	uv, ok := ve.usersVideo[id]
	if !ok {
		return false, errs.ErrNotFound()
	}

	if !uv.windowed.Load() {
		window := new(app.Window)
		window.Option(app.Title(nickname))
		uv.mu.Lock()
		uv.window = window
		uv.windowed.Store(true)
		uv.mu.Unlock()
		go uv.proccessUsersWindowWebcam()
		return true, nil
	}
	uv.mu.Lock()
	if uv.window != nil {
		uv.window.Perform(system.ActionClose)
	}
	uv.mu.Unlock()
	uv.img = nil
	return false, nil

}

func (ve *videoEngine) OnOffWebcam() (bool, error) {
	s := ve.started.Load()

	if !s == true {

		stream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
			Video: func(mtc *mediadevices.MediaTrackConstraints) {
				mtc.Width = prop.Int(320)
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
		close(ve.stopProcessWebcamChan)
		ve.stopProcessWebcamChan = nil
		if ve.webcamTrack != nil {
			if err := ve.webcamTrack.Close(); err != nil {
				ve.log.Error("failed to close webcam track", logger.Err(err))
			}
		}

		ve.webcamReader = nil

	}

	ve.userFrame.Store("")
	ve.started.Store(!s)
	return !s, nil
}

func (ve *videoEngine) SetDisconnected() {
	ve.connected.Store(false)
	ve.started.Store(false)
	ve.userFrame.Store("")
	ve.mu.Lock()
	for _, uv := range ve.usersVideo {
		uv.mu.Lock()
		if uv.window != nil {
			uv.window.Perform(system.ActionClose)
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
			frame, realese, err := ve.webcamReader.Read()
			if err != nil {
				ve.log.Error("failed to read frame", logger.Err(err))
				continue
			}

			ve.webcamBuffer.Reset()

			if err := webp.Encode(ve.webcamBuffer, frame, eOpt); err != nil {
				ve.log.Error("failed to encode frame", logger.Err(err))
				continue
			}

			realese()

			buf := ve.webcamBuffer.Bytes()

			buffer := ve.bytesBuffersPool.Get().([]byte)
			copy(buffer, buf)

			n := len(buf)

			select {
			case ve.videoFrameChan <- buffer[:n]:
				var f string
				o := ve.GetUserFrame()
				ve.mu.RLock()
				n := len(ve.usersVideo)
				ve.mu.RUnlock()
				f, err = renderLocalImg(n, frame)
				if err != nil {
					ve.log.Error("failed to render img", logger.Err(err))
					f = o
				}
				closeWaitChan()
				ve.userFrame.Store(f)
			default:
				ve.bytesBuffersPool.Put(buffer[:1350])
			}
		}

	}
}

func (uv *userVideo) proccessUsersWindowWebcam() {
	var ops op.Ops
	for {
		w := uv.window
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			uv.mu.Lock()
			if w == uv.window {
				uv.windowed.Store(false)
				uv.window = nil
			}
			uv.mu.Unlock()
			return
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			uv.mu.RLock()
			im := uv.img
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
	frame, err := renderTerminalImg(n, img)
	if err != nil {
		ve.log.Error("failed to render img", logger.Err(err))
		return
	}
	ve.mu.Lock()
	defer ve.mu.Unlock()
	uv.frame = frame

	uv.mu.Lock()
	if uv.windowed.Load() && uv.window != nil {
		uv.img = img
		uv.window.Invalidate()
	}
	uv.mu.Unlock()

}

func (ve *videoEngine) RemoveUserFromUsersVideo(id uuid.UUID) {
	ve.mu.Lock()
	defer ve.mu.Unlock()
	uv, ok := ve.usersVideo[id]
	if ok {
		uv.mu.Lock()
		if uv.window != nil {
			uv.window.Perform(system.ActionClose)
		}
		uv.mu.Unlock()
		delete(ve.usersVideo, id)
	}
}

func (ve *videoEngine) GetUsersFramesTerminal() map[uuid.UUID]string {
	ve.mu.RLock()
	defer ve.mu.RUnlock()
	frames := make(map[uuid.UUID]string, len(ve.usersVideo))
	for i, uv := range ve.usersVideo {
		frames[i] = uv.frame
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

	imageWidget.SetSizeWithCorrection(int(float64(termW)/(float64(n)*1.25)), int(float64(termH)/(float64(n)*1.25)))
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

	imageWidget.SetSizeWithCorrection(int(float64(termW)/(float64(n)*1.25)), int(float64(termH)/(float64(n)*1.25)))

	textMsg, err := imageWidget.Render()
	if err != nil {
		return "", err
	}

	return textMsg, nil
}

func (ve *videoEngine) sendVideo() {
	ve.log.Info("sending webcam")
	for {
		select {
		case <-ve.stopSendVideoChan:
			ve.log.Info("video sending stopped")
			return
		case frame := <-ve.videoFrameChan:
			if ve.connected.Load() && ve.started.Load() && ve.netw != nil {
				if err := ve.netw.SendVideoData(frame); err != nil {
					ve.log.Error("failed to send video data", logger.Err(err))
				}
			}
			ve.bytesBuffersPool.Put(frame[:1350])
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

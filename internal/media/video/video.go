package video

import (
	"aloh-tui/internal/networking"
	"aloh-tui/pkg/logger"
	"bytes"
	"image"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/term"

	"github.com/blacktop/go-termimg"
	"github.com/google/uuid"
	"gocv.io/x/gocv"
	"golang.org/x/image/webp"
)

type VideoEngine interface {
	SetNetworking(netw networking.Networking)
	SetConnected()
	SetDisconnected()
	RenderUsersVideoTerminal(id uuid.UUID, data []byte)
	GetUsersFramesTerminal() map[uuid.UUID]string
	RemoveUserFromUsersVideo(id uuid.UUID)
	OnOffWebcam() (bool, error)
	GetUserFrame() string
	IsStarted() bool
	Stop()
}

type videoEngine struct {
	webcam *gocv.VideoCapture
	netw   networking.Networking

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
	frame string
}

func NewVideoEngine(l *logger.Logger, vs VideoSetup) (VideoEngine, error) {
	log := l.AddOp("videoEngine")

	ve := &videoEngine{
		log:                   log,
		videoFrameChan:        make(chan []byte, 1),
		stopSendVideoChan:     make(chan struct{}, 1),
		stopProcessWebcamChan: make(chan struct{}, 1),
		usersVideo:            make(map[uuid.UUID]*userVideo, 3),
		bytesBuffersPool: sync.Pool{
			New: func() any {
				buf := make([]byte, 1000)
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

func (ve *videoEngine) OnOffWebcam() (bool, error) {
	s := ve.started.Load()

	if !s == true {
		webcam, err := gocv.OpenVideoCapture(0)
		if err != nil {
			ve.log.Error("failed to open webcam", logger.Err(err))
			return !s, err
		}

		ve.webcam = webcam
		ve.stopProcessWebcamChan = make(chan struct{}, 1)
		waitChan := make(chan struct{}, 1)
		go ve.processWebcam(waitChan)
		<-waitChan

	} else {
		if ve.webcam != nil {
			close(ve.stopProcessWebcamChan)
			ve.stopProcessWebcamChan = nil
			if err := ve.webcam.Close(); err != nil {
				ve.log.Error("failed to close webcam", logger.Err(err))
				return !s, err
			}
			ve.webcam = nil
		}
	}

	ve.userFrame.Store("")
	ve.started.Store(!s)
	return !s, nil
}

func (ve *videoEngine) SetDisconnected() {
	ve.connected.Store(false)
	ve.started.Store(false)
	ve.userFrame.Store("")
	clear(ve.usersVideo)
	if ve.stopProcessWebcamChan != nil {
		close(ve.stopProcessWebcamChan)
		ve.stopProcessWebcamChan = nil
	}
	if ve.webcam != nil {
		if err := ve.webcam.Close(); err != nil {
			ve.log.Error("failed to close webcam", logger.Err(err))
		}
		ve.webcam = nil
	}
}

func (ve *videoEngine) processWebcam(wc chan struct{}) {
	ve.log.Info("processing webcam")
	timer := time.NewTicker(50 * time.Millisecond)
	imge := gocv.NewMat()
	//smallMat := gocv.NewMat()
	defer imge.Close()
	closeWaitChan := sync.OnceFunc(func() {
		close(wc)
	})
	//defer smallMat.Close()
	for {
		select {
		case <-ve.stopProcessWebcamChan:
			return
		case <-timer.C:
			if ve.webcam == nil || !ve.webcam.Read(&imge) || imge.Empty() {
				continue
			}

			gocv.Resize(imge, &imge, image.Pt(320, 160), 0, 0, gocv.InterpolationLinear)

			webpBuff, err := gocv.IMEncodeWithParams(".webp", imge, []int{gocv.IMWriteWebpQuality, 1})
			if err != nil {
				continue
			}

			bytess := webpBuff.GetBytes()
			buffer := ve.bytesBuffersPool.Get().([]byte)
			copy(buffer, bytess)
			webpBuff.Close()
			n := len(bytess)

			select {
			case ve.videoFrameChan <- buffer[:n]:
				var f string
				o := ve.GetUserFrame()
				nativeImg, err := imge.ToImage()
				if err != nil {
					ve.log.Error("failed to render img", logger.Err(err))
					f = o
				} else {
					ve.mu.RLock()
					n := len(ve.usersVideo)
					ve.mu.RUnlock()
					f, err = renderLocalImg(n, nativeImg)
					if err != nil {
						ve.log.Error("failed to render img", logger.Err(err))
						f = o
					}
				}
				closeWaitChan()
				ve.userFrame.Store(f)
			default:
				ve.bytesBuffersPool.Put(buffer[:1000])
			}
		}

	}
}

func (ve *videoEngine) RenderUsersVideoTerminal(id uuid.UUID, data []byte) {
	ve.mu.Lock()
	uv, ok := ve.usersVideo[id]
	if !ok {
		uv = &userVideo{}
		ve.usersVideo[id] = uv
	}
	n := len(ve.usersVideo)
	ve.mu.Unlock()

	frame, err := renderImg(n, data)
	if err != nil {
		ve.log.Error("failed to render img", logger.Err(err))
	}
	ve.mu.Lock()
	uv.frame = frame
	ve.mu.Unlock()
}

func (ve *videoEngine) RemoveUserFromUsersVideo(id uuid.UUID) {
	ve.mu.Lock()
	defer ve.mu.Unlock()
	delete(ve.usersVideo, id)
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

func renderImg(n int, data []byte) (string, error) {
	img, err := webp.Decode(bytes.NewReader(data))
	if err != nil {
		return "", err
	}

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
			ve.bytesBuffersPool.Put(frame[:1000])
		}
	}
}

func (ve *videoEngine) Stop() {

	close(ve.stopSendVideoChan)

	if ve.webcam != nil {
		if err := ve.webcam.Close(); err != nil {
			ve.log.Error("failed to close webcam", logger.Err(err))
		}
	}

}

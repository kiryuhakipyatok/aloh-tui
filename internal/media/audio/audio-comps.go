package audio

import (
	"aloh-tui/internal/media/audio/filter"
	"aloh-tui/pkg/logger"
	"sync"
	"sync/atomic"

	"github.com/gen2brain/malgo"
	"github.com/kechako/go-speexdsp"
	rnnoise "github.com/kiryuhakipyatok/rnnoise/cmd"
	"gopkg.in/hraban/opus.v2"
)

type AudioSetup struct {
	Microphone  string
	Headphones  string
	HardDenoice bool
	FullMuted   bool
	MicMuted    bool
	SoftDenoice bool
	Aec         bool
	Filtered    bool
}

type DeviceInfo struct {
	Index      int
	Name       string
	Format     string
	Channels   uint32
	SampleRate uint32
}

type buffers struct {
	voiceBuffer []byte

	denoicedBuffer []float32
	float32Buffer  []float32

	pcmBuffer            []int16
	echolessBufferInt16s []int16
	micNativeBuffer      []int16
	workMic              []int16
	playbackNativeBuffer []int16
	workMix              []int16
	resampledWorkMix     []int16
	resampledWorkMic     []int16
	monoCaptureBuffer    []int16
}

type atmoics struct {
	mutedMicro atomic.Bool
	muted      atomic.Bool

	hardDenoiced atomic.Bool
	softDenoiced atomic.Bool

	aec atomic.Bool

	filtered atomic.Bool

	captureReady  atomic.Bool
	playbackReady atomic.Bool

	switching      atomic.Bool
	connected      atomic.Bool
	userIsSpeaking atomic.Bool
}

type devices struct {
	captureDevice  *malgo.Device
	playbackDevice *malgo.Device

	Microphones map[string]DeviceInfo
	Headphones  map[string]DeviceInfo

	CurrentMicrophone DeviceInfo
	CurrentHeadphones DeviceInfo
}

type speexdspComps struct {
	echoCanceller     *speexdsp.EchoCanceller
	preprocessor      *speexdsp.Preprocessor
	captureResampler  *speexdsp.Resampler
	playbackResampler *speexdsp.Resampler
}

type callbacks struct {
	captureCallback  malgo.DeviceCallbacks
	playbackCallback malgo.DeviceCallbacks
}

type chans struct {
	micDataChan       chan []byte
	stopSendVoiceChan chan struct{}
}

type pools struct {
	bytesBuffersPool sync.Pool
	int16BuffersPool sync.Pool
}

type filters struct {
	lowShelfFilter  *filter.BiquadFilter
	highShelfFilter *filter.BiquadFilter
}

type sounds struct {
	notificationSound []byte
	notificationBytes []byte
	notificationPos   int
}

type logComps struct {
	log         *logger.SparseLogger
	logCount    uint
	errLogCount uint
}

type mutexes struct {
	mu          sync.RWMutex
	lifecycleMu sync.Mutex
}

type usersAudio struct {
	data                 []byte
	decoder              *opus.Decoder
	isSpeaking           atomic.Bool
	playing              bool
	framesCount          uint
	muted                atomic.Bool
	volumeCoefficient    float32
	decodedBuffer        []byte
	rms                  float64
	samples              []int16
	float32Buffer        []float32
	denoicedBuffer       []float32
	personalPreprocessor *speexdsp.Preprocessor
	softDenoised         atomic.Bool
	personalHardDenoise  *rnnoise.RNNoise
	hardDenoised         atomic.Bool
}

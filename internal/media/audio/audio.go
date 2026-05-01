package audio

import (
	"aloh-tui/internal/media/audio/filter"
	"aloh-tui/internal/networking"
	"aloh-tui/internal/notifications"
	"aloh-tui/pkg/errs"
	"aloh-tui/pkg/logger"
	"encoding/binary"
	"errors"
	"math"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/gen2brain/malgo"
	"github.com/kechako/go-speexdsp"
	rnnoise "github.com/kiryuhakipyatok/rnnoise/cmd"
	"gopkg.in/hraban/opus.v2"
)

type AudioEngine interface {
	PlayNotification()
	PlayUserVoice(nickname string, userVoiceByte []byte)
	Stop()
	SetNetworking(netw networking.Networking) error
	SetConnected()
	SetDisconnected()
	ChangeMicrophone(microphone string) error
	FetchMicrophones() []malgo.DeviceInfo
	MuteUnmuteMicro() bool
	MuteUnmute() bool
	SetVolume(nickname string, vc float32)
	MuteUnmuteUser(nickname string) (bool, error)
	SetMuteState(nickname string, mute bool)
	FetchUsersMutes() map[string]struct{}
	FetchSpeakingUsers() map[string]struct{}
	UserIsSpeaking() bool
	OnOffDenoice() bool
	OnOffAEC() bool
	OnOffFilter() bool
}

type usersAudio struct {
	data              []byte
	decoder           *opus.Decoder
	isSpeaking        atomic.Bool
	playing           bool
	framesCount       uint
	muted             atomic.Bool
	volumeCoefficient float32
	decodedBuffer     []byte
}

type audioEngine struct {
	captureDevice  *malgo.Device
	playbackDevice *malgo.Device
	malgoCtx       *malgo.AllocatedContext

	mutedMicro atomic.Bool
	muted      atomic.Bool
	denoiced   atomic.Bool
	aec        atomic.Bool
	filtered   atomic.Bool

	bytesBuffersPool sync.Pool
	int16BuffersPool sync.Pool

	stopSendVocie chan struct{}

	notificationBytes []byte
	usersAudio        map[string]*usersAudio

	voiceHolder atomic.Int32
	threshold   float64

	userIsSpeaking atomic.Bool

	rnnoise *rnnoise.RNNoise

	lowShelfFilter  *filter.BiquadFilter
	highShelfFilter *filter.BiquadFilter

	notificationPos int

	mu          sync.Mutex
	lifecycleMu sync.Mutex
	switching   atomic.Bool

	netw   networking.Networking
	sounds sounds

	opusEncoder *opus.Encoder

	voiceBuffer          []byte
	denoicedBuffer       []float32
	pcmBuffer            []int16
	float32Buffer        []float32
	echolessBufferInt16s []int16
	micNativeBuffer      []int16
	workMic              []int16
	playbackNativeBuffer []int16
	workMix              []int16
	resampledWorkMix     []int16
	resampledWorkMic     []int16

	connected bool

	echoCanceller     *speexdsp.EchoCanceller
	preprocessor      *speexdsp.Preprocessor
	captureResampler  *speexdsp.Resampler
	playbackResampler *speexdsp.Resampler

	log *logger.SparseLogger

	Microphones []malgo.DeviceInfo

	micDataChan chan []byte

	captureCallback  malgo.DeviceCallbacks
	playbackCallback malgo.DeviceCallbacks

	logCount uint

	errLogCount uint
}

type sounds struct {
	notification []byte
}

const (
	frameSize        = 1920
	frameLen         = 960
	jitterSize       = 9600
	rnnoiseFrameSize = 480
)

func NewAudioEngine(l *logger.Logger, microphone string, denoice, aec bool) (AudioEngine, error) {
	log := l.AddOp("audioEngine")

	sparseLogger := l.Sparse(20)

	notificationSound := notifications.NotificationSoundBytes()

	opusEncoder, err := opus.NewEncoder(48000, 1, opus.Application(opus.AppVoIP))
	if err != nil {
		log.Error("failed to create new opus encoder", logger.Err(err))
		return nil, err
	}

	if err = opusEncoder.SetBitrate(40000); err != nil {
		log.Error("failed to set bitrate to opus encoder", logger.Err(err))
		return nil, err
	}

	rnnoise := rnnoise.NewRNNoise()

	echoCanceller := speexdsp.NewEchoCanceller(1, 1, 48000, 960, 4800)

	preprocessor := speexdsp.NewPreprocessor(48000, 960)
	preprocessor.SetEchoCanceller(echoCanceller)
	preprocessor.EnableDenoise(false)

	lowShelfFilter := filter.NewLowShelfFilter(48000, 200, 3)
	highShelfFilter := filter.NewHighShelfFilter(48000, 3500, 3)

	ae := &audioEngine{
		sounds: sounds{
			notification: notificationSound,
		},
		stopSendVocie: make(chan struct{}, 1),
		usersAudio:    make(map[string]*usersAudio, 10),
		Microphones:   make([]malgo.DeviceInfo, 0, 7),

		bytesBuffersPool: sync.Pool{
			New: func() any {
				buf := make([]byte, 1000)
				return buf
			},
		},
		int16BuffersPool: sync.Pool{
			New: func() any {
				buf := make([]int16, 4096)
				return buf
			},
		},

		lowShelfFilter:  lowShelfFilter,
		highShelfFilter: highShelfFilter,

		voiceBuffer:          make([]byte, 1000),
		denoicedBuffer:       make([]float32, frameLen),
		float32Buffer:        make([]float32, frameLen),
		pcmBuffer:            make([]int16, frameLen),
		playbackNativeBuffer: make([]int16, 0, 4096),
		micNativeBuffer:      make([]int16, 0, 4096),
		echolessBufferInt16s: make([]int16, frameLen),
		workMix:              make([]int16, 4096),
		workMic:              make([]int16, 4096),
		resampledWorkMix:     make([]int16, 4096),
		resampledWorkMic:     make([]int16, 4096),

		echoCanceller: echoCanceller,
		preprocessor:  preprocessor,
		opusEncoder:   opusEncoder,
		log:           sparseLogger,
		threshold:     150,
		rnnoise:       rnnoise,
		micDataChan:   make(chan []byte, 100),
	}

	ae.denoiced.Store(denoice)
	ae.aec.Store(aec)

	ctx, err := malgo.InitContext(audioBackends, malgo.ContextConfig{}, nil)
	if err != nil {
		log.Error("failed to init malgo context", logger.Err(err))
		return nil, err
	}

	ae.malgoCtx = ctx

	microphones, err := ctx.Devices(malgo.Capture)
	if err != nil {
		log.Error("failed to get devices", logger.Err(err))
		return nil, err
	}

	ae.Microphones = microphones

	var micId unsafe.Pointer
	if microphone != "" {
		id, err := ae.resolveCaptureDeviceByName(microphone)
		if err != nil {
			micId = nil
		} else {
			micId = id
		}
	}

	captureConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	playbackConfig := malgo.DefaultDeviceConfig(malgo.Playback)

	captureConfig.Capture.Format = malgo.FormatS16
	captureConfig.Capture.Channels = 1
	captureConfig.SampleRate = 0
	captureConfig.Capture.DeviceID = micId
	captureConfig.PeriodSizeInFrames = 960

	playbackConfig.Playback.Channels = 1
	playbackConfig.Playback.Format = malgo.FormatS16
	playbackConfig.SampleRate = 0
	playbackConfig.PeriodSizeInFrames = 960

	ae.captureCallback = ae.newCaptureCallback()
	ae.playbackCallback = ae.newPlaybackCallback()

	captureDevice, err := malgo.InitDevice(ae.malgoCtx.Context, captureConfig, ae.captureCallback)
	if err != nil {
		ae.log.Error(ae.errLogCount, "failed to init malgo capture device", logger.Err(err))
		ae.malgoCtx.Free()
		return nil, err
	}

	playbackDevice, err := malgo.InitDevice(ae.malgoCtx.Context, playbackConfig, ae.playbackCallback)
	if err != nil {
		ae.log.Error(ae.errLogCount, "failed to init malgo playback device", logger.Err(err))
		ae.malgoCtx.Free()
		return nil, err
	}

	ae.captureDevice = captureDevice
	ae.playbackDevice = playbackDevice

	captureSampleRate := int(ae.captureDevice.SampleRate())
	playbackSampleRate := int(ae.playbackDevice.SampleRate())

	captureResampler, err := speexdsp.NewResampler(1, captureSampleRate, 48000, 6)
	if err != nil {
		ae.log.Error(ae.errLogCount, "failed to create new capture resampler", logger.Err(err))
		return nil, err
	}

	playbackResampler, err := speexdsp.NewResampler(1, 48000, playbackSampleRate, 6)
	if err != nil {
		ae.log.Error(ae.errLogCount, "failed to create new playback resampler", logger.Err(err))
		return nil, err
	}

	ae.captureResampler = captureResampler
	ae.playbackResampler = playbackResampler

	if err := ae.start(); err != nil {
		ae.log.Error(ae.errLogCount, "failed to start audio engine", logger.Err(err))
		ae.malgoCtx.Free()
		return nil, err
	}

	go ae.sendVoice()

	return ae, nil
}

func (ae *audioEngine) resolveCaptureDeviceByName(name string) (unsafe.Pointer, error) {
	devices, err := ae.malgoCtx.Devices(malgo.Capture)
	if err != nil {
		return nil, err
	}
	ae.mu.Lock()
	ae.Microphones = devices
	ae.mu.Unlock()
	for i := range devices {
		if devices[i].Name() == name {
			return devices[i].ID.Pointer(), nil
		}
	}
	return nil, errors.New("selected microphone not found")
}

func (ae *audioEngine) FetchMicrophones() []malgo.DeviceInfo {
	return ae.Microphones
}

func (ae *audioEngine) ChangeMicrophone(microphone string) error {
	ae.lifecycleMu.Lock()
	defer ae.lifecycleMu.Unlock()

	ae.switching.Store(true)
	defer ae.switching.Store(false)

	var micId unsafe.Pointer
	if microphone != "" {
		id, err := ae.resolveCaptureDeviceByName(microphone)
		if err != nil {
			micId = nil
		} else {
			micId = id
		}
	}

	if ae.captureDevice != nil {
		ae.captureDevice.Stop()
		ae.captureDevice.Uninit()
		ae.mu.Lock()
		ae.captureDevice = nil
		ae.voiceHolder.Store(0)

		ae.mu.Unlock()

	}

	captureConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	captureConfig.Capture.Format = malgo.FormatS16
	captureConfig.Capture.Channels = 1
	captureConfig.SampleRate = 0
	captureConfig.Capture.DeviceID = micId

	captureDevice, err := malgo.InitDevice(ae.malgoCtx.Context, captureConfig, ae.captureCallback)
	if err != nil && captureConfig.Capture.DeviceID != nil {
		captureConfig.Capture.DeviceID = nil
		captureDevice, err = malgo.InitDevice(ae.malgoCtx.Context, captureConfig, ae.captureCallback)
	}
	if err != nil {
		return err
	}

	if err := ae.captureResampler.Close(); err != nil {
		ae.log.Error(ae.errLogCount, "failed to close old capture resampler", logger.Err(err))
		return err
	}

	captureSampleRate := int(captureDevice.SampleRate())

	captureResampler, err := speexdsp.NewResampler(1, captureSampleRate, 48000, 6)
	if err != nil {
		ae.log.Error(ae.errLogCount, "failed to create new capture resampler", logger.Err(err))
		return err
	}

	ae.mu.Lock()
	ae.captureDevice = captureDevice
	ae.captureResampler = captureResampler
	ae.mu.Unlock()

	if err := ae.start(); err != nil {
		ae.captureDevice.Uninit()
		ae.mu.Lock()
		ae.captureDevice = nil
		ae.mu.Unlock()
		return err
	}

	return nil
}

func (ae *audioEngine) MuteUnmuteMicro() bool {
	s := ae.mutedMicro.Load()
	if ae.muted.Load() && s {
		ae.muted.Store(false)
	}
	ae.mutedMicro.Store(!s)
	return !s
}

func (ae *audioEngine) MuteUnmute() bool {
	s := ae.muted.Load()
	if ae.mutedMicro.Load() && s {
		ae.mutedMicro.Store(false)
	}
	ae.muted.Store(!s)
	return !s
}

func (ae *audioEngine) OnOffDenoice() bool {
	s := ae.denoiced.Load()
	ae.denoiced.Store(!s)
	return !s
}

func (ae *audioEngine) OnOffAEC() bool {
	s := ae.aec.Load()
	ae.aec.Store(!s)
	return !s
}

func (ae *audioEngine) OnOffFilter() bool {
	s := ae.filtered.Load()
	ae.filtered.Store(!s)
	return !s
}

func (ae *audioEngine) newCaptureCallback() malgo.DeviceCallbacks {
	data := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		if ae.switching.Load() {
			return
		}

		ae.mu.Lock()
		connected := ae.connected
		netw := ae.netw
		ae.mu.Unlock()

		if pInputSamples != nil && netw != nil && connected && !ae.mutedMicro.Load() && !ae.muted.Load() {

			nativeSamples := len(pInputSamples) / 2

			if len(ae.micNativeBuffer) < nativeSamples {
				ae.micNativeBuffer = make([]int16, nativeSamples)
			}

			bytesToInt16(ae.micNativeBuffer[:nativeSamples], pInputSamples)

			outLen := int(float64(nativeSamples)*(48000.0/float64(ae.captureDevice.SampleRate()))) + 100
			if len(ae.resampledWorkMic) < outLen {
				ae.resampledWorkMic = make([]int16, outLen)
			}

			_, out, err := ae.captureResampler.ProcessInt(0, ae.micNativeBuffer[:nativeSamples], ae.resampledWorkMic)
			if err == nil {
				ae.workMic = append(ae.workMic, ae.resampledWorkMic[:out]...)
			}

			for len(ae.workMic) >= frameLen {
				chunk := ae.workMic[:frameLen]
				copy(ae.pcmBuffer, chunk)

				if ae.aec.Load() {
					ae.echoCanceller.Capture(ae.pcmBuffer, ae.echolessBufferInt16s)
					ae.preprocessor.Run(ae.echolessBufferInt16s)
					copy(ae.pcmBuffer, ae.echolessBufferInt16s)
				}

				var voiceDetected bool

				if ae.denoiced.Load() {
					int16ToFloat32(ae.pcmBuffer, ae.float32Buffer)

					for i := 0; i+rnnoiseFrameSize <= len(ae.float32Buffer); i += rnnoiseFrameSize {
						rnnFrame := ae.float32Buffer[i : i+rnnoiseFrameSize]
						outChunk := ae.denoicedBuffer[i : i+rnnoiseFrameSize]
						vad, err := ae.rnnoise.Denoise(outChunk, rnnFrame)
						if err != nil {
							ae.log.Error(0, "failed to denoise frame", logger.Err(err))
						}

						if vad > 0.50 {
							voiceDetected = true
						}
					}

					ae.float32ToInt16(ae.pcmBuffer, ae.denoicedBuffer)
				} else {
					var (
						sum      float64
						zcr      int
						lastSign bool
					)

					for i, sample := range ae.pcmBuffer {
						val := float64(sample)
						sum += val * val

						sign := sample > 0
						if i > 0 && sign != lastSign {
							zcr++
						}
						lastSign = sign
					}

					rms := math.Sqrt(sum / float64(len(ae.pcmBuffer)))

					if rms > ae.threshold || (rms > 15 && zcr > 100) {
						voiceDetected = true
					}
				}

				var isVoice bool

				if voiceDetected {
					isVoice = true
					ae.userIsSpeaking.Store(true)
					ae.voiceHolder.Store(30)
				} else {
					if ae.voiceHolder.Load() > 0 {
						isVoice = true
						ae.userIsSpeaking.Store(true)
						ae.voiceHolder.Add(-1)
					} else {
						isVoice = false
						ae.userIsSpeaking.Store(false)
					}
				}

				if !isVoice {
					copy(ae.workMic, ae.workMic[frameLen:])
					ae.workMic = ae.workMic[:len(ae.workMic)-frameLen]
					continue
				}
				if ae.filtered.Load() {
					ae.filter(ae.pcmBuffer)
				}
				n, err := ae.opusEncoder.Encode(ae.pcmBuffer, ae.voiceBuffer)
				if err != nil {
					ae.log.Error(ae.errLogCount, "failed to encode opus data", logger.Err(err))
				} else {
					packetToSend := ae.bytesBuffersPool.Get().([]byte)
					copy(packetToSend[:n], ae.voiceBuffer[:n])
					select {
					case ae.micDataChan <- packetToSend[:n]:
					default:
					}
				}

				copy(ae.workMic, ae.workMic[frameLen:])
				ae.workMic = ae.workMic[:len(ae.workMic)-frameLen]
			}
		} else {
			ae.userIsSpeaking.Store(false)
			ae.workMic = ae.workMic[:0]

		}
	}
	return malgo.DeviceCallbacks{
		Data: data,
	}
}

func (ae *audioEngine) newPlaybackCallback() malgo.DeviceCallbacks {
	data := func(pOutputSample, pInputSamples []byte, framecount uint32) {

		if pOutputSample != nil {
			for i := range pOutputSample {
				pOutputSample[i] = 0
			}

			nativeSamples := len(pOutputSample) / 2

			for len(ae.playbackNativeBuffer) < nativeSamples {
				for i := 0; i < frameLen; i++ {
					ae.workMix[i] = 0
				}
				ae.mu.Lock()
				for _, ua := range ae.usersAudio {

					if len(ua.data) == 0 {
						ua.framesCount++
						if ua.framesCount > 5 {
							ua.isSpeaking.Store(false)
						}
						ua.playing = false
						ua.data = nil
						continue
					} else {
						ua.framesCount = 0

					}

					if !ua.playing {
						if len(ua.data) >= jitterSize {
							ua.playing = true
						} else {
							ua.isSpeaking.Store(false)
							continue
						}
					}

					ua.isSpeaking.Store(true)

					uaLen := len(ua.data)

					readLen := frameSize

					if uaLen < readLen {
						readLen = uaLen
					}

					mixBytesToInt16(ae.workMix, ua.data[:readLen])

					ua.data = ua.data[readLen:]

					if len(ua.data) == 0 {
						ua.data = nil
						ua.playing = false
					}

				}

				if ae.notificationBytes != nil {
					rem := len(ae.notificationBytes) - ae.notificationPos

					readLen := frameSize
					if rem < readLen {
						readLen = rem
					}

					chunk := ae.notificationBytes[ae.notificationPos : ae.notificationPos+readLen]

					mixBytesToInt16(ae.workMix, chunk)
					ae.notificationPos += readLen

					if ae.notificationPos >= len(ae.notificationBytes) {
						ae.notificationBytes = nil
						ae.notificationPos = 0
					}
				}

				ae.mu.Unlock()

				if ae.aec.Load() {
					ae.echoCanceller.Playback(ae.workMix[:frameLen])
				}

				outLen := int(float64(frameLen)*(float64(ae.playbackDevice.SampleRate())/48000.0)) + 100
				if len(ae.resampledWorkMix) < outLen {
					ae.resampledWorkMix = make([]int16, outLen)
				}

				_, out, err := ae.playbackResampler.ProcessInt(0, ae.workMix[:frameLen], ae.resampledWorkMix)
				if err == nil {
					ae.playbackNativeBuffer = append(ae.playbackNativeBuffer, ae.resampledWorkMix[:out]...)
				}

			}

			outNativeBuffer := ae.playbackNativeBuffer[:nativeSamples]

			int16ToBytes(outNativeBuffer, pOutputSample)

			copy(ae.playbackNativeBuffer, ae.playbackNativeBuffer[nativeSamples:])
			ae.playbackNativeBuffer = ae.playbackNativeBuffer[:len(ae.playbackNativeBuffer)-nativeSamples]
		}

	}

	return malgo.DeviceCallbacks{
		Data: data,
	}
}

func setupVolume(vc float32, pcm []int16) {
	if vc == 1 {
		return
	}
	for i := range pcm {
		val := vc * float32(pcm[i])

		if val > math.MaxInt16 {
			val = math.MaxInt16
		}

		if val < math.MinInt16 {
			val = math.MinInt16
		}

		pcm[i] = int16(val)

	}
}

func (ae *audioEngine) PlayNotification() {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	ae.notificationBytes = ae.sounds.notification
	ae.notificationPos = 0
}

func (ae *audioEngine) MuteUnmuteUser(nickname string) (bool, error) {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	ua, ok := ae.usersAudio[nickname]
	if !ok {
		return false, errs.ErrNotFound
	}
	s := ua.muted.Load()
	ua.muted.Store(!s)
	return !s, nil
}

func (ae *audioEngine) SetMuteState(nickname string, mute bool) {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	ua, ok := ae.usersAudio[nickname]
	if !ok {
		opusDecoder, err := opus.NewDecoder(48000, 1)
		if err != nil {
			ae.log.Error(ae.errLogCount, "failed to create new opus decoder", logger.Err(err))
			return
		}
		ua = &usersAudio{
			data:              make([]byte, 0, 48000),
			decoder:           opusDecoder,
			volumeCoefficient: 1,
			decodedBuffer:     make([]byte, 5760),
		}

		ua.muted.Store(mute)

		ae.usersAudio[nickname] = ua

		return
	}
	ua.muted.Store(mute)
}

func (ae *audioEngine) PlayUserVoice(nickname string, userVoiceByte []byte) {
	if ae.muted.Load() || ae.switching.Load() {
		return
	}
	ae.mu.Lock()
	ua, ok := ae.usersAudio[nickname]
	if !ok {
		opusDecoder, err := opus.NewDecoder(48000, 1)
		if err != nil {
			ae.log.Error(ae.errLogCount, "failed to create new opus decoder", logger.Err(err))
			ae.mu.Unlock()
			return
		}

		ua = &usersAudio{
			data:              make([]byte, 0, 48000),
			decoder:           opusDecoder,
			volumeCoefficient: 1,
			decodedBuffer:     make([]byte, 5760),
		}

		ae.usersAudio[nickname] = ua
	}
	if ua.muted.Load() {
		ae.mu.Unlock()
		return
	}
	v := ua.volumeCoefficient
	ae.mu.Unlock()
	pcmBuffer := ae.int16BuffersPool.Get().([]int16)
	n, err := ua.decoder.Decode(userVoiceByte, pcmBuffer)
	if err != nil {
		ae.log.Error(ae.errLogCount, "failed to decode incoming opus packet", logger.Err(err))
		return
	}
	setupVolume(v, pcmBuffer[:n])
	int16ToBytes(pcmBuffer[:n], ua.decodedBuffer[:n*2])
	ae.mu.Lock()
	ae.int16BuffersPool.Put(pcmBuffer[:4096])
	ua.data = append(ua.data, ua.decodedBuffer[:n*2]...)

	if len(ua.data) > 96000 {
		ua.data = ua.data[len(ua.data)-jitterSize:]
	}
	ae.mu.Unlock()
}

func (ae *audioEngine) UserIsSpeaking() bool {
	return ae.userIsSpeaking.Load()
}

func (ae *audioEngine) SetVolume(nickname string, vc float32) {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	ua, ok := ae.usersAudio[nickname]
	if !ok {
		opusDecoder, err := opus.NewDecoder(48000, 1)
		if err != nil {
			ae.log.Error(ae.errLogCount, "failed to create new opus decoder", logger.Err(err))
			return
		}
		ua = &usersAudio{
			data:              make([]byte, 0, 48000),
			decoder:           opusDecoder,
			volumeCoefficient: vc,
			decodedBuffer:     make([]byte, 5760),
		}

		ae.usersAudio[nickname] = ua

		return
	}
	ua.volumeCoefficient = vc
}

func (ae *audioEngine) start() error {
	if ae.captureDevice != nil {
		if err := ae.captureDevice.Start(); err != nil {
			return err
		}
		ae.log.Info(0, "capture device started")
	}
	if ae.playbackDevice != nil {
		if err := ae.playbackDevice.Start(); err != nil {
			return err
		}
		ae.log.Info(0, "playback device started")
	}
	return nil
}

func (ae *audioEngine) Stop() {
	ae.lifecycleMu.Lock()
	defer ae.lifecycleMu.Unlock()
	if ae.playbackDevice != nil {
		if err := ae.playbackDevice.Stop(); err != nil {
			ae.log.Error(0, "failed to stop playback device", logger.Err(err))
		}
		ae.playbackDevice.Uninit()
		ae.playbackDevice = nil
	}

	if ae.captureDevice != nil {
		if err := ae.captureDevice.Stop(); err != nil {
			ae.log.Error(0, "failed to stop capture device", logger.Err(err))
		}
		ae.captureDevice.Uninit()
		ae.captureDevice = nil
	}

	if ae.malgoCtx != nil {
		ae.malgoCtx.Free()
		ae.malgoCtx = nil
	}

	if ae.echoCanceller != nil {
		if err := ae.echoCanceller.Close(); err != nil {
			ae.log.Error(0, "failed to close echo canceller", logger.Err(err))
		}
	}

	if ae.preprocessor != nil {
		if err := ae.preprocessor.Close(); err != nil {
			ae.log.Error(0, "failed to close preprocessor", logger.Err(err))
		}
	}

	if ae.captureResampler != nil {
		if err := ae.captureResampler.Close(); err != nil {
			ae.log.Error(0, "failed to close capture resampler", logger.Err(err))
		}
	}

	if ae.playbackResampler != nil {
		if err := ae.playbackResampler.Close(); err != nil {
			ae.log.Error(0, "failed to close playback resampler", logger.Err(err))
		}
	}

	ae.stopSendVocie <- struct{}{}
	close(ae.stopSendVocie)
	close(ae.micDataChan)

	ae.log.Info(0, "audio engine stopped")
}

func (ae *audioEngine) SetNetworking(netw networking.Networking) error {
	if netw == nil {
		return errors.New("networking is nil")
	}
	ae.mu.Lock()
	ae.netw = netw
	ae.mu.Unlock()
	return nil
}

func (ae *audioEngine) FetchSpeakingUsers() map[string]struct{} {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	speakers := make(map[string]struct{}, len(ae.usersAudio))
	for n, ua := range ae.usersAudio {
		if ua.isSpeaking.Load() {
			speakers[n] = struct{}{}
		}
	}
	return speakers
}

func (ae *audioEngine) FetchUsersMutes() map[string]struct{} {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	muters := make(map[string]struct{}, len(ae.usersAudio))
	for n, ua := range ae.usersAudio {
		if ua.muted.Load() {
			muters[n] = struct{}{}
		}
	}
	return muters
}

func (ae *audioEngine) SetConnected() {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	ae.connected = true
}

func (ae *audioEngine) SetDisconnected() {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	ae.connected = false
}

func (ae *audioEngine) sendVoice() {
	for {
		select {
		case <-ae.stopSendVocie:
			ae.log.Info(0, "stopping voice sending")
			return
		case voice := <-ae.micDataChan:
			if ae.mutedMicro.Load() {
				continue
			}

			ae.mu.Lock()
			connected := ae.connected
			netw := ae.netw
			ae.mu.Unlock()
			if connected && netw != nil {
				if err := ae.netw.SendVoiceData(voice); err != nil {
					ae.log.Error(ae.errLogCount, "failed to send voice data", logger.Err(err))
				}
			}
			ae.bytesBuffersPool.Put(voice[:1000])
		}
	}
}

func mixBytesToInt16(int16s []int16, b []byte) {
	samples := len(b) / 2
	for i := 0; i < samples && i < len(int16s); i++ {
		val := int16(binary.LittleEndian.Uint16(b[i*2 : i*2+2]))
		mixed := int32(int16s[i]) + int32(val)
		if mixed > math.MaxInt16 {
			mixed = math.MaxInt16
		} else if mixed < math.MinInt16 {
			mixed = math.MinInt16
		}
		int16s[i] = int16(mixed)
	}
}

func int16ToBytes(int16s []int16, b []byte) {
	for i := 0; i < len(int16s) && i*2+1 < len(b); i++ {
		binary.LittleEndian.PutUint16(b[i*2:i*2+2], uint16(int16s[i]))
	}
}

func bytesToInt16(int16s []int16, b []byte) {
	for i := 0; i < len(int16s) && i*2+1 < len(b); i++ {
		val := int16(binary.LittleEndian.Uint16(b[i*2 : i*2+2]))
		int16s[i] = val
	}
}

func (ae *audioEngine) float32ToInt16(ints16 []int16, floats []float32) {
	for i := 0; i < len(floats) && i < len(ints16); i++ {
		f := floats[i]
		if f > math.MaxInt16 {
			f = math.MaxInt16
		} else if f < math.MinInt16 {
			f = math.MinInt16
		}
		ints16[i] = int16(f)
	}
}

func int16ToFloat32(ints16 []int16, floats []float32) {
	for i := 0; i < len(floats) && i < len(ints16); i++ {
		intt := ints16[i]
		floats[i] = float32(intt)
	}
}

func (ae *audioEngine) filter(samples []int16) {
	for i := 0; i < len(samples); i++ {
		filtered := ae.lowShelfFilter.Process(float32(samples[i]))
		filtered = ae.highShelfFilter.Process(filtered)

		if filtered > math.MaxInt16 {
			filtered = math.MaxInt16
		} else if filtered < math.MinInt16 {
			filtered = math.MinInt16
		}

		samples[i] = int16(filtered)
	}
}

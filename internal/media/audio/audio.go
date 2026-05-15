package audio

import (
	"aloh-tui/internal/media/audio/backends"
	"aloh-tui/internal/media/audio/casters"
	"aloh-tui/internal/media/audio/filter"
	"aloh-tui/internal/networking"
	"aloh-tui/internal/notifications"
	"aloh-tui/pkg/errs"
	"aloh-tui/pkg/logger"
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
	FetchMicrophones() map[string]MicrophoneInfo
	UpdateMicrophones() error
	MuteUnmuteMicro() bool
	MuteUnmute() bool
	SetVolume(nickname string, vc float32)
	MuteUnmuteUser(nickname string) (bool, error)
	SetMuteState(nickname string, mute bool)
	FetchUsersMutes() map[string]struct{}
	FetchSpeakingUsers() map[string]struct{}
	GetCurrentMicrophone() MicrophoneInfo
	UserIsSpeaking() bool
	OnOffHardDenoice() bool
	OnOffSoftDenoice() bool
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

	leftChannel bool

	mutedMicro    atomic.Bool
	muted         atomic.Bool
	hardDenoiced  atomic.Bool
	softDenoiced  atomic.Bool
	aec           atomic.Bool
	filtered      atomic.Bool
	captureReady  atomic.Bool
	playbackReady atomic.Bool

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
	monoCaptureBuffer    []int16

	connected bool

	echoCanceller     *speexdsp.EchoCanceller
	preprocessor      *speexdsp.Preprocessor
	captureResampler  *speexdsp.Resampler
	playbackResampler *speexdsp.Resampler

	log *logger.SparseLogger

	Microphones       map[string]MicrophoneInfo
	CurrentMicrophone MicrophoneInfo
	CurrentHeadphones MicrophoneInfo

	micDataChan chan []byte

	captureCallback  malgo.DeviceCallbacks
	playbackCallback malgo.DeviceCallbacks

	logCount uint

	errLogCount uint
}

type sounds struct {
	notification []byte
}

type MicrophoneInfo struct {
	Index      int
	Name       string
	Channels   uint32
	SampleRate uint32
}

type AudioSetup struct {
	Microphone  string
	HardDenoice bool
	SoftDenoice bool
	Aec         bool
	Filtered    bool
}

const (
	frameSize        = 1920
	frameLen         = 960
	jitterSize       = 13440
	rnnoiseFrameSize = 480
)

func NewAudioEngine(l *logger.Logger, as AudioSetup) (AudioEngine, error) {
	log := l.AddOp("audioEngine")

	sparseLogger := l.Sparse(20)

	notificationSound := notifications.NotificationSoundBytes()

	opusEncoder, err := opus.NewEncoder(48000, 1, opus.Application(opus.AppVoIP))
	if err != nil {
		log.Error("failed to create new opus encoder", logger.Err(err))
		return nil, err
	}

	if err = opusEncoder.SetBitrate(64000); err != nil {
		log.Error("failed to set bitrate to opus encoder", logger.Err(err))
		return nil, err
	}

	if err = opusEncoder.SetMaxBandwidth(opus.Fullband); err != nil {
		log.Error("failed to set max bandwidth to opus encoder", logger.Err(err))
		return nil, err
	}
	if err = opusEncoder.SetInBandFEC(true); err != nil {
		log.Error("failed to set in band fec to opus encoder", logger.Err(err))
		return nil, err
	}

	rnnoise := rnnoise.NewRNNoise()

	echoCanceller := speexdsp.NewEchoCanceller(1, 1, 48000, 960, 4800)

	preprocessor := speexdsp.NewPreprocessor(48000, 960)

	if as.Aec {
		preprocessor.SetEchoCanceller(echoCanceller)
	}

	preprocessor.EnableDenoise(as.SoftDenoice)

	lowShelfFilter := filter.NewLowShelfFilter(48000, 200, -4)
	highShelfFilter := filter.NewHighShelfFilter(48000, 4500, 4)

	ae := &audioEngine{
		sounds: sounds{
			notification: notificationSound,
		},
		stopSendVocie: make(chan struct{}, 1),
		usersAudio:    make(map[string]*usersAudio, 10),

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
		monoCaptureBuffer:    make([]int16, 0, 4096),

		echoCanceller: echoCanceller,
		preprocessor:  preprocessor,
		opusEncoder:   opusEncoder,
		log:           sparseLogger,
		threshold:     100,
		rnnoise:       rnnoise,
		micDataChan:   make(chan []byte, 100),
	}

	ae.softDenoiced.Store(as.SoftDenoice)
	ae.hardDenoiced.Store(as.HardDenoice)
	ae.aec.Store(as.Aec)
	ae.filtered.Store(as.Filtered)

	ctx, err := malgo.InitContext(backends.AudioBackends, malgo.ContextConfig{}, nil)
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

	micsInfo := make(map[string]MicrophoneInfo, 0)

	var (
		ch    uint32
		micId unsafe.Pointer
	)

	for i, m := range microphones {
		di, err := ctx.DeviceInfo(malgo.Capture, m.ID, malgo.Shared)
		if err != nil {
			log.Error("failed to get devices info", logger.Err(err))
			return nil, err
		}
		log.Debug("mic's info", logger.Attr("name", m.Name()), logger.Attr("formats", di.Formats))
		format := di.Formats[0]
		mi := MicrophoneInfo{
			Index:      i,
			Name:       m.Name(),
			SampleRate: format.SampleRate,
			Channels:   format.Channels,
		}
		micsInfo[m.Name()] = mi

		if di.IsDefault == 1 {
			ae.CurrentMicrophone = mi
			ch = format.Channels
		}

		if as.Microphone != "" && microphones[i].Name() == as.Microphone {
			micId = microphones[i].ID.Pointer()
			ch = format.Channels
			ae.CurrentMicrophone = mi
			log.Debug("current mic's info", logger.Attr("name", mi.Name),
				logger.Attr("sampleRate", format.SampleRate), logger.Attr("channels", format.Channels),
				logger.Attr("format", format.Format))
		}

	}

	ae.Microphones = micsInfo

	captureConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	playbackConfig := malgo.DefaultDeviceConfig(malgo.Playback)

	captureConfig.Capture.Format = 0
	if ch > 2 {
		captureConfig.Capture.Channels = 2
	} else {
		captureConfig.Capture.Channels = 0
	}
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

	log.Debug("current capture device", logger.Attr("channels", captureDevice.CaptureChannels()),
		logger.Attr("sampleRate", captureDevice.SampleRate()), logger.Attr("format", captureDevice.CaptureFormat()))

	playbackDevice, err := malgo.InitDevice(ae.malgoCtx.Context, playbackConfig, ae.playbackCallback)
	if err != nil {
		ae.log.Error(ae.errLogCount, "failed to init malgo playback device", logger.Err(err))
		ae.malgoCtx.Free()
		return nil, err
	}

	log.Debug("current playback device", logger.Attr("channels", playbackDevice.PlaybackChannels()),
		logger.Attr("sampleRate", playbackDevice.SampleRate()), logger.Attr("format", playbackDevice.PlaybackFormat()))

	ae.captureDevice = captureDevice
	ae.playbackDevice = playbackDevice

	captureSampleRate := int(ae.captureDevice.SampleRate())
	playbackSampleRate := int(ae.playbackDevice.SampleRate())

	captureResampler, err := speexdsp.NewResampler(1, captureSampleRate, 48000, 7)
	if err != nil {
		ae.log.Error(ae.errLogCount, "failed to create new capture resampler", logger.Err(err))
		return nil, err
	}

	playbackResampler, err := speexdsp.NewResampler(1, 48000, playbackSampleRate, 7)
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
	for i := range devices {
		if devices[i].Name() == name {
			return devices[i].ID.Pointer(), nil
		}
	}
	return nil, errors.New("selected microphone not found")
}

func (ae *audioEngine) newCaptureCallback() malgo.DeviceCallbacks {
	data := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		if ae.switching.Load() {
			return
		}

		ae.captureReady.Store(true)

		ae.mu.Lock()
		connected := ae.connected
		netw := ae.netw
		ae.mu.Unlock()

		if pInputSamples != nil && netw != nil && connected && !ae.mutedMicro.Load() && !ae.muted.Load() {
			cf := ae.captureDevice.CaptureFormat()
			var nativeSamples int
			switch cf {
			case malgo.FormatU8:
				nativeSamples = len(pInputSamples)
			case malgo.FormatS16:
				nativeSamples = len(pInputSamples) / 2
			case malgo.FormatS24:
				nativeSamples = len(pInputSamples) / 3
			case malgo.FormatS32:
				nativeSamples = len(pInputSamples) / 4
			case malgo.FormatF32:
				nativeSamples = len(pInputSamples) / 4
			default:
				ae.log.Error(0, "invalid capture format", logger.Attr("format", cf))
				return
			}

			if len(ae.micNativeBuffer) < nativeSamples {
				ae.micNativeBuffer = make([]int16, nativeSamples)
			}

			switch cf {
			case malgo.FormatU8:
				casters.BytesU8ToInt16(ae.micNativeBuffer[:nativeSamples], pInputSamples)
			case malgo.FormatS16:
				casters.BytesS16ToInt16(ae.micNativeBuffer[:nativeSamples], pInputSamples)
			case malgo.FormatS24:
				casters.BytesS24ToInt16(ae.micNativeBuffer[:nativeSamples], pInputSamples)
			case malgo.FormatS32:
				casters.BytesS32ToInt16(ae.micNativeBuffer[:nativeSamples], pInputSamples)
			case malgo.FormatF32:
				casters.BytesF32ToInt16(ae.micNativeBuffer[:nativeSamples], pInputSamples)
			default:
				ae.log.Error(0, "invalid capture format", logger.Attr("format", cf))
				return
			}

			usingMicBuffer := ae.micNativeBuffer[:nativeSamples]

			if ae.captureDevice.CaptureChannels() == 2 {
				monoSamples := nativeSamples / 2
				if len(ae.monoCaptureBuffer) < monoSamples {
					ae.monoCaptureBuffer = make([]int16, monoSamples)
				}
				ae.stereoToMono(ae.micNativeBuffer[:nativeSamples], ae.monoCaptureBuffer[:monoSamples])
				usingMicBuffer = ae.monoCaptureBuffer[:monoSamples]
			}
			if ae.captureDevice.SampleRate() == 48000 {
				ae.workMic = append(ae.workMic, usingMicBuffer...)
			} else {
				outLen := int(float64(nativeSamples)*(48000.0/float64(ae.captureDevice.SampleRate()))) + 100
				if len(ae.resampledWorkMic) < outLen {
					ae.resampledWorkMic = make([]int16, outLen)
				}
				_, out, err := ae.captureResampler.ProcessInt(0, usingMicBuffer, ae.resampledWorkMic)
				if err == nil {
					ae.workMic = append(ae.workMic, ae.resampledWorkMic[:out]...)
				}
			}

			for len(ae.workMic) >= frameLen {
				chunk := ae.workMic[:frameLen]
				copy(ae.pcmBuffer, chunk)

				ae.mu.Lock()
				if ae.aec.Load() {
					if ae.playbackReady.Load() {
						ae.echoCanceller.Capture(ae.pcmBuffer, ae.echolessBufferInt16s)
						ae.preprocessor.Run(ae.echolessBufferInt16s)
						copy(ae.pcmBuffer, ae.echolessBufferInt16s)
					} else {
						ae.preprocessor.Run(ae.pcmBuffer)
					}
				} else if ae.softDenoiced.Load() {
					ae.preprocessor.Run(ae.pcmBuffer)
				}

				ae.mu.Unlock()

				var voiceDetected bool

				if ae.hardDenoiced.Load() {
					casters.Int16ToFloat32(ae.pcmBuffer, ae.float32Buffer)

					for i := 0; i+rnnoiseFrameSize <= len(ae.float32Buffer); i += rnnoiseFrameSize {
						rnnFrame := ae.float32Buffer[i : i+rnnoiseFrameSize]
						outChunk := ae.denoicedBuffer[i : i+rnnoiseFrameSize]
						vad, err := ae.rnnoise.Denoise(outChunk, rnnFrame)
						if err != nil {
							ae.log.Error(0, "failed to denoise frame", logger.Err(err))
						}

						if vad > 0.40 {
							voiceDetected = true
						}
					}

					casters.Float32ToInt16(ae.pcmBuffer, ae.denoicedBuffer)
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
						if ae.userIsSpeaking.Load() {
							if err := ae.opusEncoder.Reset(); err != nil {
								ae.log.Error(0, "failed to reset opus state", logger.Err(err))
							}
							ae.userIsSpeaking.Store(false)
						}

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
			ae.playbackReady.Store(true)
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

					casters.MixBytesToInt16(ae.workMix, ua.data[:readLen])

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

					casters.MixBytesToInt16(ae.workMix, chunk)
					ae.notificationPos += readLen

					if ae.notificationPos >= len(ae.notificationBytes) {
						ae.notificationBytes = nil
						ae.notificationPos = 0
					}
				}

				if ae.aec.Load() && ae.captureReady.Load() {
					ae.echoCanceller.Playback(ae.workMix[:frameLen])
				}
				ae.mu.Unlock()

				if ae.playbackDevice.SampleRate() == 48000 {
					ae.playbackNativeBuffer = append(ae.playbackNativeBuffer, ae.workMix[:frameLen]...)
				} else {
					outLen := int(float64(frameLen)*(float64(ae.playbackDevice.SampleRate())/48000.0)) + 100
					if len(ae.resampledWorkMix) < outLen {
						ae.resampledWorkMix = make([]int16, outLen)
					}

					_, out, err := ae.playbackResampler.ProcessInt(0, ae.workMix[:frameLen], ae.resampledWorkMix)
					if err == nil {
						ae.playbackNativeBuffer = append(ae.playbackNativeBuffer, ae.resampledWorkMix[:out]...)
					}
				}

			}

			outNativeBuffer := ae.playbackNativeBuffer[:nativeSamples]

			casters.Int16ToBytes(outNativeBuffer, pOutputSample)

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
	casters.Int16ToBytes(pcmBuffer[:n], ua.decodedBuffer[:n*2])
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

func (ae *audioEngine) FetchMicrophones() map[string]MicrophoneInfo {
	return ae.Microphones
}

func (ae *audioEngine) ChangeMicrophone(microphone string) error {
	ae.lifecycleMu.Lock()
	defer ae.lifecycleMu.Unlock()

	ae.switching.Store(true)
	defer ae.switching.Store(false)

	ae.captureReady.Store(false)

	var (
		micId unsafe.Pointer
		ch    uint32
	)

	if microphone != "" {
		id, err := ae.resolveCaptureDeviceByName(microphone)
		if err != nil {
			micId = nil
		} else {
			micId = id
		}
	}

	micInfo, ok := ae.Microphones[microphone]
	if ok && micId != nil {
		ae.CurrentMicrophone = micInfo
		ch = micInfo.Channels
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
	captureConfig.Capture.Format = 0
	if ch > 2 {
		captureConfig.Capture.Channels = 2
	} else {
		captureConfig.Capture.Channels = 0
	}
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

	captureResampler, err := speexdsp.NewResampler(1, captureSampleRate, 48000, 7)
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
	ae.muted.Store(false)
	ae.mutedMicro.Store(!s)
	return !s
}

func (ae *audioEngine) MuteUnmute() bool {
	s := ae.muted.Load()
	ae.mutedMicro.Store(false)
	ae.muted.Store(!s)
	return !s
}

func (ae *audioEngine) OnOffHardDenoice() bool {
	h := ae.hardDenoiced.Load()
	ae.hardDenoiced.Store(!h)
	return !h
}

func (ae *audioEngine) OnOffSoftDenoice() bool {
	s := ae.softDenoiced.Load()
	ae.mu.Lock()
	if ae.preprocessor != nil {
		ae.preprocessor.EnableDenoise(!s)
	}
	ae.mu.Unlock()
	ae.softDenoiced.Store(!s)
	return !s
}

func (ae *audioEngine) GetCurrentMicrophone() MicrophoneInfo {
	return ae.CurrentMicrophone
}

func (ae *audioEngine) OnOffAEC() bool {
	s := ae.aec.Load()
	ae.aec.Store(!s)
	ae.mu.Lock()
	if s {
		ae.preprocessor.SetEchoCanceller(nil)
	} else {
		ae.preprocessor.SetEchoCanceller(ae.echoCanceller)
	}
	ae.mu.Unlock()
	return !s
}

func (ae *audioEngine) OnOffFilter() bool {
	s := ae.filtered.Load()
	ae.filtered.Store(!s)
	return !s
}

func (ae *audioEngine) UpdateMicrophones() error {
	microphones, err := ae.malgoCtx.Devices(malgo.Capture)
	if err != nil {
		ae.log.Error(0, "failed to get devices", logger.Err(err))
		return err
	}

	micsInfo := make(map[string]MicrophoneInfo, 0)

	for i, m := range microphones {
		di, err := ae.malgoCtx.DeviceInfo(malgo.Capture, m.ID, malgo.Shared)
		if err != nil {
			ae.log.Error(0, "failed to get devices info", logger.Err(err))
			return err
		}
		format := di.Formats[0]
		micsInfo[m.Name()] = MicrophoneInfo{
			Index:      i,
			Name:       m.Name(),
			SampleRate: format.SampleRate,
			Channels:   format.Channels,
		}
	}
	ae.Microphones = micsInfo
	return nil
}

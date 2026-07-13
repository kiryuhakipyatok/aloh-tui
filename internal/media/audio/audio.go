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
	"github.com/google/uuid"
	"github.com/kechako/go-speexdsp"
	rnnoise "github.com/kiryuhakipyatok/rnnoise/cmd"
	"gopkg.in/hraban/opus.v2"
)

type AudioEngine interface {
	PlayNotification()
	PlayUserVoice(id uuid.UUID, userVoiceByte []byte)
	Stop()
	SetNetworking(netw networking.Networking) error
	SetConnected()
	SetDisconnected() error
	ChangeMicrophone(microphone string) error
	ChangeHeadphones(headphone string) error
	FetchMicrophones() map[string]DeviceInfo
	FetchHeadphones() map[string]DeviceInfo
	UpdateMicrophones() error
	UpdateHeadphones() error
	MuteUnmuteMicro() bool
	MuteUnmute() bool
	SetVolume(id uuid.UUID, vc float32)
	MuteUnmuteUser(id uuid.UUID) (bool, error)
	SetMuteState(id uuid.UUID, mute bool)
	FetchUsersMutes() map[uuid.UUID]struct{}
	FetchSpeakingUsers() map[uuid.UUID]float64
	GetCurrentMicrophone() DeviceInfo
	GetCurrentHeadphones() DeviceInfo
	UserIsSpeaking() bool
	OnOffHardDenoice() bool
	OnOffSoftDenoice() bool
	OnOffAEC() bool
	OnOffFilter() bool
	OnOffUsersHardDenoise(id uuid.UUID, state bool) error
	OnOffUsersSoftDenoise(id uuid.UUID, state bool) error
	RemoveFromUsersAudio(id uuid.UUID) error
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
	usersAudio        map[uuid.UUID]*usersAudio

	voiceHolder atomic.Int32
	threshold   float64

	userIsSpeaking atomic.Bool

	rnnoise *rnnoise.RNNoise

	lowShelfFilter  *filter.BiquadFilter
	highShelfFilter *filter.BiquadFilter

	notificationPos int

	mu          sync.RWMutex
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

	connected atomic.Bool

	echoCanceller     *speexdsp.EchoCanceller
	preprocessor      *speexdsp.Preprocessor
	captureResampler  *speexdsp.Resampler
	playbackResampler *speexdsp.Resampler

	log *logger.SparseLogger

	Microphones       map[string]DeviceInfo
	Headphones        map[string]DeviceInfo
	CurrentMicrophone DeviceInfo
	CurrentHeadphones DeviceInfo

	micDataChan chan []byte

	captureCallback  malgo.DeviceCallbacks
	playbackCallback malgo.DeviceCallbacks

	logCount uint

	errLogCount uint
}

type sounds struct {
	notification []byte
}

type DeviceInfo struct {
	Index      int
	Name       string
	Channels   uint32
	SampleRate uint32
	Format     string
}

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

const (
	frameSize        = 1920
	frameLen         = 960
	jitterSize       = 13440
	rnnoiseFrameSize = 480
)

const (
	int16f   = "int16"
	float32f = "float32"
	int8f    = "int8"
	int24f   = "int24"
	int32f   = "int32"
	unk      = "unknown"
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
		usersAudio:    make(map[uuid.UUID]*usersAudio, 10),

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
		threshold:     150,
		rnnoise:       rnnoise,
		micDataChan:   make(chan []byte, 100),
	}

	ae.softDenoiced.Store(as.SoftDenoice)
	ae.hardDenoiced.Store(as.HardDenoice)
	ae.aec.Store(as.Aec)
	ae.filtered.Store(as.Filtered)
	ae.muted.Store(as.FullMuted)
	ae.mutedMicro.Store(as.MicMuted)

	ctx, err := malgo.InitContext(backends.AudioBackends, malgo.ContextConfig{}, nil)
	if err != nil {
		log.Error("failed to init malgo context", logger.Err(err))
		return nil, err
	}

	ae.malgoCtx = ctx

	microphones, err := ctx.Devices(malgo.Capture)
	if err != nil {
		log.Error("failed to get capture devices", logger.Err(err))
		return nil, err
	}

	heeadphones, err := ctx.Devices(malgo.Playback)
	if err != nil {
		log.Error("failed to get playback devices", logger.Err(err))
		return nil, err
	}

	micsInfo := make(map[string]DeviceInfo, 0)
	headsInfo := make(map[string]DeviceInfo, 0)

	var (
		ch     uint32
		micId  unsafe.Pointer
		headId unsafe.Pointer
	)

	for i, m := range microphones {
		di, err := ctx.DeviceInfo(malgo.Capture, m.ID, malgo.Shared)
		if err != nil {
			log.Error("failed to get capture devices info", logger.Err(err))
			return nil, err
		}
		log.Debug("mic's info", logger.Attr("name", m.Name()), logger.Attr("formats", di.Formats))
		format := di.Formats[0]
		mi := DeviceInfo{
			Index:      i,
			Name:       m.Name(),
			SampleRate: format.SampleRate,
			Channels:   format.Channels,
		}

		switch format.Format {
		case malgo.FormatU8:
			mi.Format = int8f
		case malgo.FormatS16:
			mi.Format = int16f
		case malgo.FormatS24:
			mi.Format = int24f
		case malgo.FormatS32:
			mi.Format = int32f
		case malgo.FormatF32:
			mi.Format = float32f
		default:
			mi.Format = unk
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

	for i, m := range heeadphones {
		di, err := ctx.DeviceInfo(malgo.Playback, m.ID, malgo.Shared)
		if err != nil {
			log.Error("failed to get playback devices info", logger.Err(err))
			return nil, err
		}
		log.Debug("headphones's info", logger.Attr("name", m.Name()), logger.Attr("formats", di.Formats))
		format := di.Formats[0]
		hi := DeviceInfo{
			Index:      i,
			Name:       m.Name(),
			SampleRate: format.SampleRate,
			Channels:   format.Channels,
		}

		switch format.Format {
		case malgo.FormatU8:
			hi.Format = int8f
		case malgo.FormatS16:
			hi.Format = int16f
		case malgo.FormatS24:
			hi.Format = int24f
		case malgo.FormatS32:
			hi.Format = int32f
		case malgo.FormatF32:
			hi.Format = float32f
		default:
			hi.Format = unk
		}
		headsInfo[m.Name()] = hi

		if di.IsDefault == 1 {
			ae.CurrentHeadphones = hi
			ch = format.Channels
		}

		if as.Headphones != "" && heeadphones[i].Name() == as.Headphones {
			headId = heeadphones[i].ID.Pointer()
			ch = format.Channels
			ae.CurrentHeadphones = hi
			log.Debug("current headphones's info", logger.Attr("name", hi.Name),
				logger.Attr("sampleRate", format.SampleRate), logger.Attr("channels", format.Channels),
				logger.Attr("format", format.Format))
		}

	}

	ae.Microphones = micsInfo
	ae.Headphones = headsInfo

	captureConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	playbackConfig := malgo.DefaultDeviceConfig(malgo.Playback)

	captureConfig.Capture.Format = 0
	if ch > 2 {
		captureConfig.Capture.Channels = 2
	}
	captureConfig.SampleRate = 0
	captureConfig.Capture.DeviceID = micId
	captureConfig.PeriodSizeInFrames = 960

	playbackConfig.Playback.Channels = 1
	playbackConfig.Playback.Format = malgo.FormatS16
	playbackConfig.SampleRate = 0
	playbackConfig.PeriodSizeInFrames = 960
	playbackConfig.Playback.DeviceID = headId

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

func (ae *audioEngine) resolveDeviceByName(name string, typee malgo.DeviceType) (unsafe.Pointer, error) {
	devices, err := ae.malgoCtx.Devices(typee)
	if err != nil {
		return nil, err
	}
	for i := range devices {
		if devices[i].Name() == name {
			return devices[i].ID.Pointer(), nil
		}
	}
	return nil, errs.ErrNotFound()
}

func (ae *audioEngine) newCaptureCallback() malgo.DeviceCallbacks {
	data := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		if ae.switching.Load() {
			return
		}

		ae.captureReady.Store(true)

		ae.mu.RLock()
		netw := ae.netw
		ae.mu.RUnlock()

		if pInputSamples != nil && netw != nil && ae.connected.Load() && !ae.mutedMicro.Load() && !ae.muted.Load() {
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

			if cap(ae.micNativeBuffer) < nativeSamples {
				ae.micNativeBuffer = make([]int16, nativeSamples+300)
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

						if vad > 0.45 {
							voiceDetected = true
						}
					}

					casters.Float32ToInt16(ae.pcmBuffer, ae.denoicedBuffer)
				} else {
					rms, zcr := getRmsAndZcr(ae.pcmBuffer)

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
						ae.bytesBuffersPool.Put(packetToSend[:1000])
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
		if ae.switching.Load() {
			return
		}

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
						if ua.framesCount <= 5 {
							plcBuffer := ae.int16BuffersPool.Get().([]int16)

							n, err := ua.decoder.Decode(nil, plcBuffer)
							if err == nil && n > 0 {
								setupVolume(ua.volumeCoefficient, plcBuffer[:n])
								casters.MixToInt16(ae.workMix, plcBuffer[:n])
							}
							ae.int16BuffersPool.Put(plcBuffer[:4096])
						} else {
							ua.isSpeaking.Store(false)
							ua.playing = false
						}
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

					casters.BytesS16ToInt16(ua.samples, ua.data[:readLen])

					if ua.hardDenoised.Load() {
						casters.Int16ToFloat32(ua.samples, ua.float32Buffer)

						for i := 0; i+rnnoiseFrameSize <= len(ua.float32Buffer); i += rnnoiseFrameSize {
							rnnFrame := ua.float32Buffer[i : i+rnnoiseFrameSize]
							outChunk := ua.denoicedBuffer[i : i+rnnoiseFrameSize]
							_, err := ua.personalHardDenoise.Denoise(outChunk, rnnFrame)
							if err != nil {
								ae.log.Error(0, "failed to denoise frame", logger.Err(err))
							}
						}

						casters.Float32ToInt16(ua.samples, ua.denoicedBuffer)
					}

					if ua.softDenoised.Load() {
						ua.personalPreprocessor.Run(ua.samples)
					}

					ua.rms = getRms(ua.samples)

					casters.MixToInt16(ae.workMix, ua.samples)

					if readLen < len(ua.data) {
						copied := copy(ua.data, ua.data[readLen:])
						ua.data = ua.data[:copied]
					} else {
						ua.data = ua.data[:0]
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
					if cap(ae.resampledWorkMix) < outLen {
						ae.resampledWorkMix = make([]int16, outLen+300)
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
				ae.bytesBuffersPool.Put(voice[:1000])
				continue
			}

			ae.mu.RLock()
			netw := ae.netw
			ae.mu.RUnlock()
			if ae.connected.Load() && netw != nil {
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

func (ae *audioEngine) MuteUnmuteUser(id uuid.UUID) (bool, error) {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	ua, ok := ae.usersAudio[id]
	if !ok {
		return false, errs.ErrNotFound()
	}
	s := ua.muted.Load()
	ua.muted.Store(!s)
	return !s, nil
}

func (ae *audioEngine) SetMuteState(id uuid.UUID, mute bool) {
	ae.mu.Lock()
	ua, ok := ae.usersAudio[id]
	if ok {
		ua.muted.Store(mute)
		ae.mu.Unlock()
		return
	}
	ae.mu.Unlock()

	opusDecoder, err := opus.NewDecoder(48000, 1)
	if err != nil {
		ae.log.Error(ae.errLogCount, "failed to create new opus decoder", logger.Err(err))
		return
	}
	newUa := &usersAudio{
		data:              make([]byte, 0, 48000),
		float32Buffer:     make([]float32, frameLen),
		denoicedBuffer:    make([]float32, frameLen),
		decoder:           opusDecoder,
		volumeCoefficient: 1,
		decodedBuffer:     make([]byte, 5760),
		samples:           make([]int16, frameLen),
	}

	ae.mu.Lock()
	defer ae.mu.Unlock()
	if existingUa, ok := ae.usersAudio[id]; ok {
		existingUa.muted.Store(mute)
	} else {
		ae.usersAudio[id] = newUa
	}
}

func (ae *audioEngine) PlayUserVoice(id uuid.UUID, userVoiceByte []byte) {
	if ae.muted.Load() || ae.switching.Load() {
		return
	}
	ae.mu.Lock()
	ua, ok := ae.usersAudio[id]
	if !ok {
		opusDecoder, err := opus.NewDecoder(48000, 1)
		if err != nil {
			ae.log.Error(ae.errLogCount, "failed to create new opus decoder", logger.Err(err))
			ae.mu.Unlock()
			return
		}

		ua = &usersAudio{
			data:              make([]byte, 0, 48000),
			float32Buffer:     make([]float32, frameLen),
			denoicedBuffer:    make([]float32, frameLen),
			decoder:           opusDecoder,
			volumeCoefficient: 1,
			decodedBuffer:     make([]byte, 5760),
			samples:           make([]int16, frameLen),
		}

		ae.usersAudio[id] = ua
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
	ae.mu.Lock()
	casters.Int16ToBytes(pcmBuffer[:n], ua.decodedBuffer[:n*2])

	ua.data = append(ua.data, ua.decodedBuffer[:n*2]...)
	ae.int16BuffersPool.Put(pcmBuffer[:4096])

	if len(ua.data) > 96000 {
		ua.data = ua.data[len(ua.data)-jitterSize:]
	}
	ae.mu.Unlock()
}

func (ae *audioEngine) UserIsSpeaking() bool {
	return ae.userIsSpeaking.Load()
}

func (ae *audioEngine) SetVolume(id uuid.UUID, vc float32) {
	ae.mu.Lock()
	ua, ok := ae.usersAudio[id]
	if ok {
		ua.volumeCoefficient = vc
		ae.mu.Unlock()
		return
	}
	ae.mu.Unlock()

	opusDecoder, err := opus.NewDecoder(48000, 1)
	if err != nil {
		ae.log.Error(ae.errLogCount, "failed to create new opus decoder", logger.Err(err))
		return
	}
	newUa := &usersAudio{
		data:              make([]byte, 0, 48000),
		float32Buffer:     make([]float32, frameLen),
		denoicedBuffer:    make([]float32, frameLen),
		decoder:           opusDecoder,
		volumeCoefficient: vc,
		decodedBuffer:     make([]byte, 5760),
		samples:           make([]int16, frameLen),
	}

	ae.mu.Lock()
	defer ae.mu.Unlock()
	if existingUa, ok := ae.usersAudio[id]; ok {
		existingUa.volumeCoefficient = vc
	} else {
		ae.usersAudio[id] = newUa
	}
}

func (ae *audioEngine) OnOffUsersHardDenoise(id uuid.UUID, state bool) error {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	ua, ok := ae.usersAudio[id]
	if !ok {
		return errs.ErrNotFound()
	}
	//s := ua.hardDenoised.Load()
	if state && ua.personalHardDenoise == nil {
		ua.personalHardDenoise = rnnoise.NewRNNoise()
	} else if !state && ua.personalHardDenoise != nil {
		if err := ua.personalHardDenoise.Close(); err != nil {
			return err
		}
		ua.personalHardDenoise = nil
	}
	ua.hardDenoised.Store(state)
	return nil
}

func (ae *audioEngine) OnOffUsersSoftDenoise(id uuid.UUID, state bool) error {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	ua, ok := ae.usersAudio[id]
	if !ok {
		return errs.ErrNotFound()
	}
	//s := ua.softDenoised.Load()
	if state && ua.personalPreprocessor == nil {
		ua.personalPreprocessor = speexdsp.NewPreprocessor(48000, 960)
		ua.personalPreprocessor.EnableDenoise(true)
		ua.personalPreprocessor.SetEchoCanceller(nil)
	} else if !state && ua.personalPreprocessor != nil {
		if err := ua.personalPreprocessor.Close(); err != nil {
			return err
		}
		ua.personalPreprocessor = nil
	}
	ua.softDenoised.Store(state)

	return nil
}

func (ae *audioEngine) RemoveFromUsersAudio(id uuid.UUID) error {
	ua, ok := ae.usersAudio[id]
	if !ok {
		return errs.ErrNotFound()
	}
	if ua.personalPreprocessor != nil {
		if err := ua.personalPreprocessor.Close(); err != nil {
			return err
		}
	}
	if ua.personalHardDenoise != nil {
		if err := ua.personalHardDenoise.Close(); err != nil {
			return err
		}
	}
	delete(ae.usersAudio, id)
	return nil
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

func (ae *audioEngine) FetchSpeakingUsers() map[uuid.UUID]float64 {
	ae.mu.RLock()
	defer ae.mu.RUnlock()
	speakers := make(map[uuid.UUID]float64, len(ae.usersAudio))
	for n, ua := range ae.usersAudio {
		if ua.isSpeaking.Load() {
			speakers[n] = ua.rms
		}
	}
	return speakers
}

func (ae *audioEngine) FetchUsersMutes() map[uuid.UUID]struct{} {
	ae.mu.RLock()
	defer ae.mu.RUnlock()
	muters := make(map[uuid.UUID]struct{}, len(ae.usersAudio))
	for n, ua := range ae.usersAudio {
		if ua.muted.Load() {
			muters[n] = struct{}{}
		}
	}
	return muters
}

func (ae *audioEngine) SetConnected() {
	ae.connected.Store(true)
}

func (ae *audioEngine) SetDisconnected() error {
	ae.connected.Store(false)
	ae.voiceHolder.Store(0)
	ae.userIsSpeaking.Store(false)
	// ae.micNativeBuffer = ae.micNativeBuffer[:0]
	// ae.workMic = ae.workMic[:0]
	// ae.resampledWorkMic = ae.resampledWorkMic[:0]
	// ae.monoCaptureBuffer = ae.monoCaptureBuffer
	// ae.mu.Lock()
	// defer ae.mu.Unlock()
	if ae.opusEncoder != nil {
		if err := ae.opusEncoder.Reset(); err != nil {
			return err
		}
	}

	ae.mu.Lock()
	for _, ua := range ae.usersAudio {
		ua.data = ua.data[:0]
		ua.playing = false
		ua.framesCount = 0
		ua.isSpeaking.Store(false)

		if ua.personalPreprocessor != nil {
			if err := ua.personalPreprocessor.Close(); err != nil {
				return err
			}
		}
		if ua.personalHardDenoise != nil {
			if err := ua.personalHardDenoise.Close(); err != nil {
				return err
			}
		}
	}

	clear(ae.usersAudio)

	ae.mu.Unlock()
	return nil
}

func (ae *audioEngine) FetchMicrophones() map[string]DeviceInfo {
	ae.mu.RLock()
	mics := ae.Microphones
	ae.mu.RUnlock()
	return mics
}

func (ae *audioEngine) FetchHeadphones() map[string]DeviceInfo {
	ae.mu.RLock()
	heads := ae.Headphones
	ae.mu.RUnlock()
	return heads
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
		id, err := ae.resolveDeviceByName(microphone, malgo.Capture)
		if err != nil {
			micId = nil
		} else {
			micId = id
		}
	}

	ae.mu.RLock()
	micInfo, ok := ae.Microphones[microphone]
	if ok && micId != nil {
		ch = micInfo.Channels
	}
	ae.mu.RUnlock()

	captureConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	captureConfig.Capture.Format = 0
	if ch > 2 {
		captureConfig.Capture.Channels = 2
	}
	captureConfig.SampleRate = 0
	captureConfig.Capture.DeviceID = micId

	newCaptureDevice, err := malgo.InitDevice(ae.malgoCtx.Context, captureConfig, ae.captureCallback)
	if err != nil && captureConfig.Capture.DeviceID != nil {
		captureConfig.Capture.DeviceID = nil
		newCaptureDevice, err = malgo.InitDevice(ae.malgoCtx.Context, captureConfig, ae.captureCallback)
	}
	if err != nil {
		return err
	}

	captureSampleRate := int(newCaptureDevice.SampleRate())

	newCaptureResampler, err := speexdsp.NewResampler(1, captureSampleRate, 48000, 7)
	if err != nil {
		ae.log.Error(ae.errLogCount, "failed to create new capture resampler", logger.Err(err))
		newCaptureDevice.Uninit()
		return err
	}

	ae.mu.Lock()
	oldCaptureDevice := ae.captureDevice
	oldCaptureResampler := ae.captureResampler
	oldMicrophone := ae.CurrentMicrophone

	if ok && micId != nil {
		ae.CurrentMicrophone = micInfo
	}

	ae.captureDevice = newCaptureDevice
	ae.captureResampler = newCaptureResampler

	if oldCaptureDevice != nil {
		oldCaptureDevice.Stop()
	}

	ae.voiceHolder.Store(0)
	ae.workMic = ae.workMic[:0]
	ae.micNativeBuffer = ae.micNativeBuffer[:0]
	ae.mu.Unlock()

	if err := newCaptureDevice.Start(); err != nil {
		ae.log.Error(ae.errLogCount, "failed to start new microphone, rolling back", logger.Err(err))

		newCaptureDevice.Uninit()
		newCaptureResampler.Close()

		ae.mu.Lock()
		ae.captureDevice = oldCaptureDevice
		ae.captureResampler = oldCaptureResampler
		ae.CurrentMicrophone = oldMicrophone
		ae.mu.Unlock()

		if oldCaptureDevice != nil {
			if startErr := oldCaptureDevice.Start(); startErr != nil {
				ae.log.Error(ae.errLogCount, "failed to restart old microphone after rollback", logger.Err(startErr))
			}
		}

		return err
	}

	if oldCaptureDevice != nil {
		oldCaptureDevice.Uninit()
	}
	if oldCaptureResampler != nil {
		oldCaptureResampler.Close()
	}

	return nil
}

func (ae *audioEngine) ChangeHeadphones(headphones string) error {
	ae.lifecycleMu.Lock()
	defer ae.lifecycleMu.Unlock()

	ae.switching.Store(true)
	defer ae.switching.Store(false)

	ae.playbackReady.Store(false)

	var (
		headsId unsafe.Pointer
	)

	if headphones != "" {
		id, err := ae.resolveDeviceByName(headphones, malgo.Playback)
		if err != nil {
			headsId = nil
		} else {
			headsId = id
		}
	}

	playbackConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	playbackConfig.Playback.Format = malgo.FormatS16
	playbackConfig.Playback.Channels = 1
	playbackConfig.SampleRate = 0
	playbackConfig.Playback.DeviceID = headsId

	newPlaybackDevice, err := malgo.InitDevice(ae.malgoCtx.Context, playbackConfig, ae.playbackCallback)
	if err != nil && playbackConfig.Playback.DeviceID != nil {
		playbackConfig.Playback.DeviceID = nil
		newPlaybackDevice, err = malgo.InitDevice(ae.malgoCtx.Context, playbackConfig, ae.playbackCallback)
	}
	if err != nil {
		return err
	}

	playbackSampleRate := int(newPlaybackDevice.SampleRate())

	newPlaybackResampler, err := speexdsp.NewResampler(1, playbackSampleRate, 48000, 7)
	if err != nil {
		ae.log.Error(ae.errLogCount, "failed to create new playback resampler", logger.Err(err))
		newPlaybackDevice.Uninit()
		return err
	}

	ae.mu.Lock()
	oldPlaybackDevice := ae.playbackDevice
	oldPlaybackResampler := ae.playbackResampler
	oldHeadphones := ae.CurrentHeadphones

	headsInfo, ok := ae.Headphones[headphones]
	if ok && headsId != nil {
		ae.CurrentHeadphones = headsInfo
	}

	ae.playbackDevice = newPlaybackDevice
	ae.playbackResampler = newPlaybackResampler

	if oldPlaybackDevice != nil {
		oldPlaybackDevice.Stop()
	}

	ae.playbackNativeBuffer = ae.playbackNativeBuffer[:0]
	ae.mu.Unlock()

	if err := newPlaybackDevice.Start(); err != nil {
		ae.log.Error(ae.errLogCount, "failed to start new headphones, rolling back", logger.Err(err))

		newPlaybackDevice.Uninit()
		newPlaybackResampler.Close()

		ae.mu.Lock()
		ae.playbackDevice = oldPlaybackDevice
		ae.playbackResampler = oldPlaybackResampler
		ae.CurrentHeadphones = oldHeadphones
		ae.mu.Unlock()

		if oldPlaybackDevice != nil {
			if startErr := oldPlaybackDevice.Start(); startErr != nil {
				ae.log.Error(ae.errLogCount, "failed to restart old headphones after rollback", logger.Err(startErr))
			}
		}

		return err
	}

	if oldPlaybackDevice != nil {
		oldPlaybackDevice.Uninit()
	}
	if oldPlaybackResampler != nil {
		oldPlaybackResampler.Close()
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

func (ae *audioEngine) GetCurrentMicrophone() DeviceInfo {
	ae.mu.RLock()
	curMic := ae.CurrentMicrophone
	ae.mu.RUnlock()
	return curMic
}

func (ae *audioEngine) GetCurrentHeadphones() DeviceInfo {
	ae.mu.RLock()
	curH := ae.CurrentHeadphones
	ae.mu.RUnlock()
	return curH
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
		ae.log.Error(0, "failed to get capture devices", logger.Err(err))
		return err
	}

	micsInfo := make(map[string]DeviceInfo, 0)

	for i, m := range microphones {
		di, err := ae.malgoCtx.DeviceInfo(malgo.Capture, m.ID, malgo.Shared)
		if err != nil {
			ae.log.Error(0, "failed to get capture devices info", logger.Err(err))
			return err
		}
		format := di.Formats[0]
		mi := DeviceInfo{
			Index:      i,
			Name:       m.Name(),
			SampleRate: format.SampleRate,
			Channels:   format.Channels,
		}

		switch format.Format {
		case malgo.FormatU8:
			mi.Format = int8f
		case malgo.FormatS16:
			mi.Format = int16f
		case malgo.FormatS24:
			mi.Format = int24f
		case malgo.FormatS32:
			mi.Format = int32f
		case malgo.FormatF32:
			mi.Format = float32f
		default:
			mi.Format = unk
		}
		micsInfo[m.Name()] = mi
	}
	ae.mu.Lock()
	ae.Microphones = micsInfo
	ae.mu.Unlock()
	return nil
}

func (ae *audioEngine) UpdateHeadphones() error {
	headphones, err := ae.malgoCtx.Devices(malgo.Playback)
	if err != nil {
		ae.log.Error(0, "failed to get playback devices", logger.Err(err))
		return err
	}

	headsInfo := make(map[string]DeviceInfo, 0)

	for i, m := range headphones {
		di, err := ae.malgoCtx.DeviceInfo(malgo.Playback, m.ID, malgo.Shared)
		if err != nil {
			ae.log.Error(0, "failed to get playback devices info", logger.Err(err))
			return err
		}
		format := di.Formats[0]
		hi := DeviceInfo{
			Index:      i,
			Name:       m.Name(),
			SampleRate: format.SampleRate,
			Channels:   format.Channels,
		}

		switch format.Format {
		case malgo.FormatU8:
			hi.Format = int8f
		case malgo.FormatS16:
			hi.Format = int16f
		case malgo.FormatS24:
			hi.Format = int24f
		case malgo.FormatS32:
			hi.Format = int32f
		case malgo.FormatF32:
			hi.Format = float32f
		default:
			hi.Format = unk
		}
		headsInfo[m.Name()] = hi
	}
	ae.mu.Lock()
	ae.Headphones = headsInfo
	ae.mu.Unlock()
	return nil
}

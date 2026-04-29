package audio

import (
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
	device   *malgo.Device
	malgoCtx *malgo.AllocatedContext

	mutedMicro atomic.Bool
	muted      atomic.Bool
	denoiced   atomic.Bool
	aec        atomic.Bool

	notificationBytes []byte
	usersAudio        map[string]*usersAudio

	voiceHolder int
	threshold   float64

	userIsSpeaking atomic.Bool

	rnnoise *rnnoise.RNNoise

	notificationPos int

	mu          sync.Mutex
	lifecycleMu sync.Mutex
	switching   atomic.Bool

	//filter *filter.BiquadFilter

	netw   networking.Networking
	sounds sounds

	opusEncoder *opus.Encoder

	microphoneBuffer []byte
	voiceBuffer      []byte
	denoicedBuffer   []float32
	pcmBuffer        []int16
	float32Buffer    []float32
	refBuffer        []int16
	frameBytesInt16s []int16
	//frameResampledInt16s []int16
	echolessBufferInt16s []int16

	playbackBuffer []byte
	// echollesBuffer       []byte
	//monoBuffer       []byte

	connected bool

	echoCanceller *speexdsp.EchoCanceller
	preprocessor  *speexdsp.Preprocessor
	//micResampler  *speexdsp.Resampler
	//headResampler *speexdsp.Resampler

	log *logger.SparseLogger

	Microphones []malgo.DeviceInfo

	micDataChan chan []byte

	callbacks malgo.DeviceCallbacks

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

	if err = opusEncoder.SetBitrate(64000); err != nil {
		log.Error("failed to set bitrate to opus encoder", logger.Err(err))
		return nil, err
	}

	// f := filter.NewHighPassFilter(48000, 70.0)

	rnnoise := rnnoise.NewRNNoise()

	echoCanceller := speexdsp.NewEchoCanceller(1, 1, 48000, 960, 9600)

	preprocessor := speexdsp.NewPreprocessor(48000, 960)
	preprocessor.SetEchoCanceller(echoCanceller)
	preprocessor.EnableDenoise(false)

	ae := &audioEngine{
		sounds: sounds{
			notification: notificationSound,
		},
		usersAudio:       make(map[string]*usersAudio, 10),
		microphoneBuffer: make([]byte, 0, 4096),
		playbackBuffer:   make([]byte, 7680),
		Microphones:      make([]malgo.DeviceInfo, 0, 10),
		voiceBuffer:      make([]byte, 1000),
		denoicedBuffer:   make([]float32, frameLen),
		float32Buffer:    make([]float32, frameLen),
		pcmBuffer:        make([]int16, frameLen),
		echoCanceller:    echoCanceller,
		refBuffer:        make([]int16, frameLen),
		preprocessor:     preprocessor,
		// frameResampledInt16s: make([]int16, frameLen),
		frameBytesInt16s:     make([]int16, frameLen),
		echolessBufferInt16s: make([]int16, frameLen),
		//monoBuffer:       make([]byte, 960),
		opusEncoder: opusEncoder,
		//filter:      f,
		log:         sparseLogger,
		threshold:   150,
		rnnoise:     rnnoise,
		micDataChan: make(chan []byte, 100),
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

	for _, m := range ae.Microphones {
		di, err := ctx.DeviceInfo(malgo.Capture, m.ID, malgo.Shared)
		if err != nil {
			log.Error("failed to get devices info", logger.Err(err))
			return nil, err
		}
		log.Info("m", di.Name(), di.Formats, di.FormatCount)
	}

	var micId unsafe.Pointer
	if microphone != "" {
		id, err := ae.resolveCaptureDeviceByName(microphone)
		if err != nil {
			micId = nil
		} else {
			micId = id
		}
	}

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Duplex)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 1
	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = 1
	deviceConfig.SampleRate = 48000
	deviceConfig.Capture.DeviceID = micId
	// deviceConfig.PeriodSizeInFrames = 960

	callbacks := ae.dataCallback()

	ae.callbacks = callbacks

	device, err := malgo.InitDevice(ae.malgoCtx.Context, deviceConfig, callbacks)
	if err != nil {
		ae.log.Error(ae.errLogCount, "failed to init malgo device", logger.Err(err))
		ae.malgoCtx.Free()
		return nil, err
	}

	//sr := int(device.SampleRate())

	// micResampler, err := speexdsp.NewResampler(1, sr, 48000, 5)
	// if err != nil {
	// 	ae.log.Error(ae.errLogCount, "failed to create new micResampler", logger.Err(err))
	// 	return nil, err
	// }

	// headResampler, err := speexdsp.NewResampler(1, 48000, sr, 5)
	// if err != nil {
	// 	ae.log.Error(ae.errLogCount, "failed to create new headResampler", logger.Err(err))
	// 	return nil, err
	// }
	// ae.headResampler = headResampler
	// ae.micResampler = micResampler
	ae.device = device

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

	var deviceID unsafe.Pointer
	if microphone != "" {
		id, err := ae.resolveCaptureDeviceByName(microphone)
		if err != nil {
			deviceID = nil
		} else {
			deviceID = id
		}
	}

	if ae.device != nil {
		ae.device.Stop()
		ae.device.Uninit()
		ae.mu.Lock()
		ae.device = nil
		ae.microphoneBuffer = ae.microphoneBuffer[:0]
		ae.voiceHolder = 0

		// for _, ua := range ae.usersAudio {
		// 	ua.isSpeaking.Store(false)
		// 	ua.playing = false
		// 	ua.framesCount = 0
		// 	ua.data = nil
		// }

		ae.mu.Unlock()

	}

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Duplex)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 1
	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = 1
	deviceConfig.SampleRate = 48000
	deviceConfig.Capture.DeviceID = deviceID
	// deviceConfig.PeriodSizeInFrames = 960

	callbacks := ae.dataCallback()

	dev, err := malgo.InitDevice(ae.malgoCtx.Context, deviceConfig, callbacks)
	if err != nil && deviceConfig.Capture.DeviceID != nil {
		deviceConfig.Capture.DeviceID = nil
		dev, err = malgo.InitDevice(ae.malgoCtx.Context, deviceConfig, callbacks)
	}
	if err != nil {
		return err
	}

	// if err := ae.micResampler.Close(); err != nil {
	// 	ae.log.Error(ae.errLogCount, "failed to close old micResampler", logger.Err(err))
	// 	return err
	// }

	// sr := int(dev.SampleRate())

	// micResampler, err := speexdsp.NewResampler(1, sr, 48000, 5)
	// if err != nil {
	// 	ae.log.Error(ae.errLogCount, "failed to create new micResampler", logger.Err(err))
	// 	return err
	// }

	// ae.micResampler = micResampler

	ae.mu.Lock()
	ae.device = dev
	ae.callbacks = callbacks
	ae.mu.Unlock()
	if err := ae.start(); err != nil {
		ae.device.Uninit()
		ae.mu.Lock()
		ae.device = nil
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

func (ae *audioEngine) dataCallback() malgo.DeviceCallbacks {
	data := func(pOutputSample, pInputSamples []byte, framecount uint32) {

		if ae.switching.Load() {
			return
		}

		ae.mu.Lock()
		connected := ae.connected
		netw := ae.netw
		ae.mu.Unlock()

		if pOutputSample != nil {
			for i := range pOutputSample {
				pOutputSample[i] = 0
			}

			ae.mu.Lock()

			//monoLen := len(pOutputSample) / 2

			// if cap(ae.monoBuffer) < monoLen {
			// 	ae.monoBuffer = make([]byte, monoLen)
			// }

			// ae.monoBuffer = ae.monoBuffer[:monoLen]

			// for i := range ae.monoBuffer {
			// 	ae.monoBuffer[i] = 0
			// }
			// out := make([]byte, len(pOutputSample))
			if len(ae.usersAudio) > 0 && !ae.muted.Load() {
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

					// logNick := logger.Attr("nickname", nickname)
					// logReadLen := logger.Attr("len-output-samples", readLen)
					// logUserDataLen := logger.Attr("user-data", logNick)

					ua.isSpeaking.Store(true)

					uaLen := len(ua.data)

					readLen := len(pOutputSample)

					if uaLen < readLen {
						readLen = uaLen
					}

					//ae.log.Info(ae.logCount, "playing users audio", logNick, logReadLen, logUserDataLen)

					chunk := ua.data[:readLen]

					mixPCM16(pOutputSample, chunk)

					ua.data = ua.data[readLen:]

					if len(ua.data) == 0 {
						ua.data = nil
						ua.playing = false
					}

				}
			}

			if ae.notificationBytes != nil {
				rem := len(ae.notificationBytes) - ae.notificationPos

				readLen := len(pOutputSample)
				if rem < readLen {
					readLen = rem
				}
				// logRem := logger.Attr("remain", rem)
				// logNeed := logger.Attr("need", readLen)
				// ae.log.Info(ae.logCount, "playing notification sound", logRem, logNeed)

				chunk := ae.notificationBytes[ae.notificationPos : ae.notificationPos+readLen]

				mixPCM16(pOutputSample, chunk)

				ae.notificationPos += readLen

				if ae.notificationPos >= len(ae.notificationBytes) {
					ae.notificationBytes = nil
					ae.notificationPos = 0
				}
			}
			ae.mu.Unlock()
			if ae.aec.Load() {
				ae.playbackBuffer = append(ae.playbackBuffer, pOutputSample...)
			}

			//ae.refBuffer = append(ae.refBuffer, pOutputSample...)
			//bytesToInt16(ae.refBuffer, pOutputSample)
			//monoToStereo(pOutputSample, ae.monoBuffer)
			// copy(pOutputSample, out)
		}

		if pInputSamples != nil && netw != nil && connected && !ae.mutedMicro.Load() && !ae.muted.Load() {
			// logFramecount := logger.Attr("framecount", framecount)
			// logLenInputSamples := logger.Attr("len-input-samples", len(pInputSamples))

			// ae.log.Info(ae.logCount, "microphone data processing", logFramecount, logLenInputSamples)

			ae.microphoneBuffer = append(ae.microphoneBuffer, pInputSamples...)

			for len(ae.microphoneBuffer) >= frameSize {

				frameBytes := ae.microphoneBuffer[:frameSize]

				// _, _, err := ae.micResampler.ProcessInt(0, ae.frameBytesInt16s, ae.frameResampledInt16s)
				// if err != nil {
				// 	ae.log.Error(0, "failed to resample mic data", logger.Err(err))
				// 	continue
				// }

				if ae.aec.Load() {
					if len(ae.playbackBuffer) >= frameSize {
						playbackBytes := ae.playbackBuffer[:frameSize]
						bytesToInt16(ae.refBuffer, playbackBytes)
						ae.playbackBuffer = ae.playbackBuffer[frameSize:]
					} else {
						for i := range ae.refBuffer {
							ae.refBuffer[i] = 0
						}
						ae.playbackBuffer = ae.playbackBuffer[:0]
					}
					// //refBytes := ae.refBuffer[:frameSize]
					bytesToInt16(ae.frameBytesInt16s, frameBytes)
					ae.echoCanceller.Cancellation(ae.frameBytesInt16s, ae.refBuffer, ae.echolessBufferInt16s)
					ae.preprocessor.Run(ae.echolessBufferInt16s)
					int16ToBytes(ae.echolessBufferInt16s, frameBytes)
				}
				//frameBytes = clear

				var voiceDetected bool

				if ae.denoiced.Load() {
					ae.bytesToFilteredFloat32(ae.float32Buffer, frameBytes)

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

					float32ToInt16(ae.pcmBuffer, ae.denoicedBuffer)
				} else {
					ae.bytesToFilteredInt16(ae.pcmBuffer, frameBytes)
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
					ae.voiceHolder = 30
				} else {
					if ae.voiceHolder > 0 {
						isVoice = true
						ae.userIsSpeaking.Store(true)
						ae.voiceHolder--
					} else {
						isVoice = false
						ae.userIsSpeaking.Store(false)
					}
				}

				if !isVoice {
					ae.microphoneBuffer = ae.microphoneBuffer[frameSize:]
					continue
				}

				n, err := ae.opusEncoder.Encode(ae.pcmBuffer, ae.voiceBuffer)
				if err != nil {
					ae.log.Error(ae.errLogCount, "failed to encode opus data", logger.Err(err))
				} else {
					packetToSend := make([]byte, n)
					copy(packetToSend, ae.voiceBuffer[:n])
					select {
					case ae.micDataChan <- packetToSend:
					default:
						// ae.log.Warn(ae.errLogCount, "dropping audio data")
					}
				}

				ae.microphoneBuffer = ae.microphoneBuffer[frameSize:]
			}

		} else {
			ae.userIsSpeaking.Store(false)
			ae.playbackBuffer = ae.playbackBuffer[:0]
			ae.microphoneBuffer = ae.microphoneBuffer[:0]
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
			decodedBuffer:     make([]byte, 1920),
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
			decodedBuffer:     make([]byte, 1920),
		}

		ae.usersAudio[nickname] = ua
	}
	if ua.muted.Load() {
		ae.mu.Unlock()
		return
	}
	v := ua.volumeCoefficient
	ae.mu.Unlock()
	pcmBuffer := make([]int16, 5760)
	n, err := ua.decoder.Decode(userVoiceByte, pcmBuffer)
	if err != nil {
		ae.log.Error(ae.errLogCount, "failed to decode incoming opus packet", logger.Err(err))
		return
	}
	setupVolume(v, pcmBuffer)
	int16ToBytes(pcmBuffer[:n], ua.decodedBuffer[:n*2])
	ae.mu.Lock()
	ua.data = append(ua.data, ua.decodedBuffer...)

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
			decodedBuffer:     make([]byte, 1920),
		}

		ae.usersAudio[nickname] = ua

		return
	}
	ua.volumeCoefficient = vc
}

func (ae *audioEngine) start() error {
	if ae.device != nil {
		if err := ae.device.Start(); err != nil {
			return err
		}
		ae.log.Info(0, "audio engine started")
	}
	return nil
}

func (ae *audioEngine) Stop() {
	ae.lifecycleMu.Lock()
	defer ae.lifecycleMu.Unlock()

	if ae.device != nil {
		if err := ae.device.Stop(); err != nil {
			ae.log.Error(0, "failed to stop malgo device", logger.Err(err))
		}
		ae.device.Uninit()
		ae.device = nil
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

	// if ae.micResampler != nil {
	// 	if err := ae.micResampler.Close(); err != nil {
	// 		ae.log.Error(0, "failed to close mic resampler", logger.Err(err))
	// 	}
	// }

	// if ae.headResampler != nil {
	// 	if err := ae.headResampler.Close(); err != nil {
	// 		ae.log.Error(0, "failed to close head resampler", logger.Err(err))
	// 	}
	// }

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
		voice := <-ae.micDataChan
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
	}
}

func mixPCM16(dst, src []byte) {
	for i := 0; i < len(src)-1 && i < len(dst)-1; i += 2 {
		s1 := int16(binary.LittleEndian.Uint16(dst[i:]))
		s2 := int16(binary.LittleEndian.Uint16(src[i:]))

		mixed := int32(s1) + int32(s2)
		if mixed > math.MaxInt16 {
			mixed = math.MaxInt16
		} else if mixed < math.MinInt16 {
			mixed = math.MinInt16
		}

		binary.LittleEndian.PutUint16(dst[i:], uint16(int16(mixed)))
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

func (ae *audioEngine) bytesToFilteredInt16(int16s []int16, b []byte) {
	for i := 0; i < len(int16s) && i*2+1 < len(b); i++ {
		val := int16(binary.LittleEndian.Uint16(b[i*2 : i*2+2]))
		int16s[i] = val
	}
}

func (ae *audioEngine) bytesToFilteredFloat32(float32s []float32, b []byte) {
	for i := 0; i < len(float32s) && i*2+1 < len(b); i++ {
		val := int16(binary.LittleEndian.Uint16(b[i*2 : i*2+2]))
		float32s[i] = float32(val)
	}
}

func float32ToInt16(ints16 []int16, floats []float32) {
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

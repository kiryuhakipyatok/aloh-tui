package audio

import (
	"aloh-tui/internal/media"
	"aloh-tui/internal/media/audio/backends"
	"aloh-tui/internal/media/audio/filter"
	"aloh-tui/internal/networking"
	"aloh-tui/internal/notifications"
	"aloh-tui/pkg/logger"
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
	Players
	Changers
	Fetchers
	Geters
	Updaters
	Seters
	OnOffers
	MuteUnmuters
	Checkers
	Removers
	Stop()
}

type audioEngine struct {
	malgoCtx *malgo.AllocatedContext

	leftChannel bool

	usersAudio map[uuid.UUID]*usersAudio

	voiceHolder atomic.Int32
	threshold   float64

	netw        networking.Networking
	opusEncoder *opus.Encoder
	rnnoise     *rnnoise.RNNoise

	sessionId atomic.Int32

	buffers
	atmoics
	devices
	speexdspComps
	callbacks
	chans
	pools
	filters
	sounds
	logComps
	mutexes
}

type userVoice struct {
	data      []byte
	sessionId int32
}

func NewAudioEngine(l *logger.Logger, as AudioSetup) (AudioEngine, error) {
	log := l.AddOp("audioEngine")

	sparseLogger := l.Sparse(20)

	notificationSound := notifications.NotificationSoundBytes()

	opusEncoder, err := opus.NewEncoder(sampleRate, 1, opus.Application(opus.AppVoIP))
	if err != nil {
		log.Error("failed to create new opus encoder", logger.Err(err))
		return nil, err
	}

	if err = opusEncoder.SetBitrate(bitrate); err != nil {
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

	echoCanceller := speexdsp.NewEchoCanceller(mono, mono, sampleRate, frameLen, 4800)

	preprocessor := speexdsp.NewPreprocessor(sampleRate, frameLen)

	if as.Aec {
		preprocessor.SetEchoCanceller(echoCanceller)
	}

	preprocessor.EnableDenoise(as.SoftDenoice)

	lowShelfFilter := filter.NewLowShelfFilter(sampleRate, 200, -4)
	highShelfFilter := filter.NewHighShelfFilter(sampleRate, 4500, 4)

	ae := &audioEngine{
		usersAudio: make(map[uuid.UUID]*usersAudio, 10),

		opusEncoder: opusEncoder,

		threshold: 150,
		rnnoise:   rnnoise,

		sounds: sounds{
			notificationSound: notificationSound,
		},

		buffers: buffers{
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
		},

		chans: chans{
			stopSendVoiceChan: make(chan struct{}, 1),
			micDataChan:       make(chan userVoice, 100),
		},

		pools: pools{
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
		},

		filters: filters{
			lowShelfFilter:  lowShelfFilter,
			highShelfFilter: highShelfFilter,
		},

		speexdspComps: speexdspComps{
			echoCanceller: echoCanceller,
			preprocessor:  preprocessor,
		},

		logComps: logComps{
			log: sparseLogger,
		},
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

	micsInfo := make(map[string]media.Device, 0)
	headsInfo := make(map[string]media.Device, 0)

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
	captureConfig.PeriodSizeInFrames = frameLen

	playbackConfig.Playback.Channels = 1
	playbackConfig.Playback.Format = malgo.FormatS16
	playbackConfig.SampleRate = 0
	playbackConfig.PeriodSizeInFrames = frameLen
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

	captureResampler, err := speexdsp.NewResampler(1, captureSampleRate, sampleRate, 7)
	if err != nil {
		ae.log.Error(ae.errLogCount, "failed to create new capture resampler", logger.Err(err))
		return nil, err
	}

	playbackResampler, err := speexdsp.NewResampler(1, sampleRate, playbackSampleRate, 7)
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
		case <-ae.stopSendVoiceChan:
			ae.log.Info(0, "stopping voice sending")
			return
		case voice := <-ae.micDataChan:
			sessionId := ae.sessionId.Load()
			if !ae.connected.Load() {
				ae.bytesBuffersPool.Put(voice.data[:1000])
				continue
			}

			if ae.mutedMicro.Load() {
				ae.bytesBuffersPool.Put(voice.data[:1000])
				continue
			}

			ae.mu.RLock()
			netw := ae.netw
			isConn := ae.connected.Load()

			ae.mu.RUnlock()
			if isConn && netw != nil && sessionId == voice.sessionId {
				ae.log.Info(0, "sended")
				if err := ae.netw.SendVoiceData(voice.data); err != nil {
					ae.log.Error(ae.errLogCount, "failed to send voice data", logger.Err(err))
				}
			}
			ae.bytesBuffersPool.Put(voice.data[:1000])
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

	close(ae.stopSendVoiceChan)
	close(ae.micDataChan)

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

	if ae.preprocessor != nil {
		if err := ae.preprocessor.Close(); err != nil {
			ae.log.Error(0, "failed to close preprocessor", logger.Err(err))
		}
		ae.preprocessor = nil
	}

	if ae.echoCanceller != nil {
		if err := ae.echoCanceller.Close(); err != nil {
			ae.log.Error(0, "failed to close echo canceller", logger.Err(err))
		}
		ae.echoCanceller = nil
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

	ae.log.Info(0, "audio engine stopped")
}

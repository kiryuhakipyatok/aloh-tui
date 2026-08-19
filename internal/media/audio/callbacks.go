package audio

import (
	"aloh-tui/internal/media/audio/casters"
	"aloh-tui/pkg/logger"

	"github.com/gen2brain/malgo"
)

func (ae *audioEngine) newCaptureCallback() malgo.DeviceCallbacks {
	data := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		if ae.switching.Load() {
			return
		}

		ae.captureReady.Store(true)

		if pInputSamples != nil && ae.connected.Load() && !ae.mutedMicro.Load() && !ae.muted.Load() {
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
			if ae.captureDevice.SampleRate() == sampleRate {
				ae.workMic = append(ae.workMic, usingMicBuffer...)
			} else {
				outLen := int(float64(nativeSamples)*(sampleRate/float64(ae.captureDevice.SampleRate()))) + 100
				if len(ae.resampledWorkMic) < outLen {
					ae.resampledWorkMic = make([]int16, outLen)
				}
				_, out, err := ae.captureResampler.ProcessInt(0, usingMicBuffer, ae.resampledWorkMic)
				if err == nil {
					ae.workMic = append(ae.workMic, ae.resampledWorkMic[:out]...)
				}
			}

			for len(ae.workMic) >= frameLen {
				if !ae.connected.Load() {
					ae.workMic = ae.workMic[:0]
					break
				}

				chunk := ae.workMic[:frameLen]
				copy(ae.pcmBuffer, chunk)

				//ae.mu.Lock()
				if ae.isAECActive() {
					ae.aecDiff.Add(-1)
					ae.echoCanceller.Capture(ae.pcmBuffer, ae.echolessBufferInt16s)
					ae.preprocessor.Run(ae.echolessBufferInt16s)
					copy(ae.pcmBuffer, ae.echolessBufferInt16s)
				} else if ae.softDenoiced.Load() {
					ae.preprocessor.Run(ae.pcmBuffer)
				}
				// if ae.aec.Load() {
				// 	if ae.playbackReady.Load() {
				// 		ae.log.Info(0, "aec capture len", len(ae.pcmBuffer))
				// 		ae.echoCanceller.Capture(ae.pcmBuffer, ae.echolessBufferInt16s)
				// 		ae.preprocessor.Run(ae.echolessBufferInt16s)
				// 		copy(ae.pcmBuffer, ae.echolessBufferInt16s)
				// 	} else {
				// 		ae.preprocessor.Run(ae.pcmBuffer)
				// 	}
				// } else if ae.softDenoiced.Load() {
				// 	ae.preprocessor.Run(ae.pcmBuffer)
				// }

				//	ae.mu.Unlock()

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
					sid := ae.sessionId.Load()
					uv := userVoice{
						data:      packetToSend[:n],
						sessionId: sid,
					}
					ae.log.Info(0, "sended in callback", sid)
					select {
					case ae.micDataChan <- uv:
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
			clear(pOutputSample)

			nativeSamples := len(pOutputSample) / 2
			//var wg sync.WaitGroup
			for len(ae.playbackNativeBuffer) < nativeSamples {
				for i := 0; i < frameLen; i++ {
					ae.workMix[i] = 0
				}

				ae.mu.RLock()
				usersAudio := ae.usersAudio
				ae.mu.RUnlock()
				for _, ua := range usersAudio {
					//	wg.Go(func() {
					ua.mu.Lock()
					for i := 0; i < frameLen; i++ {
						ua.workMix[i] = 0
					}
					if len(ua.data) == 0 {
						ua.framesCount++
						if ua.framesCount <= 5 {
							plcBuffer := ae.int16BuffersPool.Get().([]int16)

							n, err := ua.decoder.Decode(nil, plcBuffer)
							if err == nil && n > 0 {
								setupVolume(ua.volumeCoefficient, plcBuffer[:n])
								casters.MixToInt16(ua.workMix, plcBuffer[:n])
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
							ua.mu.Unlock()
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

					casters.MixToInt16(ua.workMix, ua.samples)

					if readLen < len(ua.data) {
						copied := copy(ua.data, ua.data[readLen:])
						ua.data = ua.data[:copied]
					} else {
						ua.data = ua.data[:0]
						ua.playing = false
					}
					ua.mu.Unlock()
					//})

				}

				//wg.Wait()

				for _, ua := range usersAudio {
					casters.MixToInt16(ae.workMix, ua.workMix)
				}

				ae.mu.Lock()
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
				if ae.isAECActive() {

					if ae.aecDiff.Load() < 4 {

						ae.echoCanceller.Playback(ae.workMix[:frameLen])
						ae.aecDiff.Add(1)
					}

				}

				if ae.playbackDevice.SampleRate() == sampleRate {
					ae.playbackNativeBuffer = append(ae.playbackNativeBuffer, ae.workMix[:frameLen]...)
				} else {
					outLen := int(float64(frameLen)*(float64(ae.playbackDevice.SampleRate())/sampleRate)) + 100
					if cap(ae.resampledWorkMix) < outLen {
						ae.resampledWorkMix = make([]int16, outLen+300)
					}

					_, out, err := ae.playbackResampler.ProcessInt(0, ae.workMix[:frameLen], ae.resampledWorkMix)
					if err == nil {
						ae.playbackNativeBuffer = append(ae.playbackNativeBuffer, ae.resampledWorkMix[:out]...)
					}
				}

				ae.mu.Unlock()

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

func (ae *audioEngine) isAECActive() bool {
	return ae.aec.Load() &&
		ae.connected.Load() &&
		!ae.mutedMicro.Load() &&
		!ae.muted.Load() &&
		ae.captureReady.Load() &&
		ae.playbackReady.Load()
}

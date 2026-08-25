package video

import (
	"aloh-tui/pkg/logger"
)

type Sender interface {
	SendWebcamKeyFrame() error
	SendScreenKeyFrame() error
}

func (ve *videoEngine) SendWebcamKeyFrame() error {
	return ve.sendKeyFrame(WEBCAM)
}

func (ve *videoEngine) SendScreenKeyFrame() error {
	return ve.sendKeyFrame(SCREEN)
}

func (ve *videoEngine) sendKeyFrame(typee uint) error {
	var vd *videoDevice

	switch typee {
	case WEBCAM:
		vd = ve.webcam
	case SCREEN:
		vd = ve.screen
	}
	ve.mu.RLock()
	kfc := vd.keyFrameCtrl
	ve.mu.RUnlock()
	if kfc != nil {
		if err := kfc.ForceKeyFrame(); err != nil {
			ve.log.Error("failed to force key frame", logger.Err(err))
			return err
		}
	}

	return nil
}

func (ve *videoEngine) sendVideo(typee uint) {
	log := setupLog(typee, ve.log)
	log.Info("sending video")

	var vd *videoDevice

	switch typee {
	case WEBCAM:
		vd = ve.webcam
	case SCREEN:
		vd = ve.screen
	}

	for {
		select {
		case <-ve.stopSendChan:
			log.Info("video sending stopped")
			return
		case buffer := <-vd.bufferChan:
			data := buffer.Bytes()
			if len(data) == 0 {
				log.Error("zero data")
				continue
			}
			ve.mu.RLock()
			netw := ve.netw
			ve.mu.RUnlock()
			if ve.connected.Load() && vd.started.Load() && netw != nil {
				switch typee {
				case WEBCAM:
					if err := netw.SendWebcamData(data); err != nil {
						log.Error("failed to send data", logger.Err(err))
					}
				case SCREEN:
					if err := netw.SendScreenData(data); err != nil {
						ve.log.Error("failed to send screen data", logger.Err(err))
					}
				}

			}
			buffer.Reset()
			data = nil
			vd.bytesBuffersPool.Put(buffer)
		}
	}
}

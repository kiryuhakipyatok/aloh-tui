package video

type Checker interface {
	IsWebcamStarted() bool
	IsScreenStarted() bool
}

func (ve *videoEngine) IsWebcamStarted() bool {
	if ve.webcam != nil {
		return ve.webcam.started.Load()
	}
	return false
}

func (ve *videoEngine) IsScreenStarted() bool {
	if ve.screen != nil {
		return ve.screen.started.Load()
	}
	return false
}

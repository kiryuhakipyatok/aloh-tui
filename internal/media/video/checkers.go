package video

type Checker interface {
	IsWebcamStarted() bool
	IsScreenStarted() bool
}

func (ve *videoEngine) IsWebcamStarted() bool {
	return ve.webcam.started.Load()
}

func (ve *videoEngine) IsScreenStarted() bool {
	return ve.screen.started.Load()
}

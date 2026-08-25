package video

type Switcher interface {
	SwitchOnScreenTab(res bool)
	SwitchOnWebcamTab(res bool)
}

func (ve *videoEngine) SwitchOnWebcamTab(res bool) {
	ve.onWebcamTab.Store(res)
}

func (ve *videoEngine) SwitchOnScreenTab(res bool) {
	ve.onScreenTab.Store(res)
}

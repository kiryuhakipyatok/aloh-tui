package audio

import "github.com/kechako/go-speexdsp"

type captureResampleNode struct {
	captureResampler *speexdsp.Resampler
}

type playbackResamplerNode struct {
	playbackResampler *speexdsp.Resampler
}

func (crn *captureResampleNode) Process(buffer []int16){
	
}

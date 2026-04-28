package filter

import (
	"math"
)

type BiquadFilter struct {
	b0, b1, b2 float64 
	a1, a2     float64 
	x1, x2     float64 
	y1, y2     float64
}


func NewHighPassFilter(sampleRate, cutoffFreq float64) *BiquadFilter {
	const Q = 0.7071

	w0 := 2.0 * math.Pi * cutoffFreq / sampleRate
	alpha := math.Sin(w0) / (2.0 * Q)
	cosW0 := math.Cos(w0)

	b0 := (1.0 + cosW0) / 2.0
	b1 := -(1.0 + cosW0)
	b2 := (1.0 + cosW0) / 2.0
	a0 := 1.0 + alpha
	a1 := -2.0 * cosW0
	a2 := 1.0 - alpha

	return &BiquadFilter{
		b0: b0 / a0,
		b1: b1 / a0,
		b2: b2 / a0,
		a1: a1 / a0,
		a2: a2 / a0,
	}
}

func (f *BiquadFilter) Process(inSample float32) float32 {
	x0 := float64(inSample)

	y0 := f.b0*x0 + f.b1*f.x1 + f.b2*f.x2 - f.a1*f.y1 - f.a2*f.y2

	f.x2 = f.x1
	f.x1 = x0
	f.y2 = f.y1
	f.y1 = y0

	return float32(y0)
}
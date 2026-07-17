package audio

import (
	"aloh-tui/pkg/errs"
	"math"
	"unsafe"

	"github.com/gen2brain/malgo"
)

func (ae *audioEngine) filter(samples []int16) {
	for i := 0; i < len(samples); i++ {
		filtered := ae.lowShelfFilter.Process(float32(samples[i]))
		filtered = ae.highShelfFilter.Process(filtered)

		if filtered > math.MaxInt16 {
			filtered = math.MaxInt16
		} else if filtered < math.MinInt16 {
			filtered = math.MinInt16
		}

		samples[i] = int16(filtered)
	}
}

func (ae *audioEngine) stereoToMono(s []int16, m []int16) {
	frames := len(s) / 2

	var (
		lsum      float64
		lzcr      int
		llastSign bool

		rsum      float64
		rzcr      int
		rlastSign bool
	)

	for i := 0; i < frames; i++ {
		lsample := s[i*2]
		rsample := s[i*2+1]

		lval := float64(lsample)
		rval := float64(rsample)

		lsum += lval * lval
		rsum += rval * rval

		lsign := lsample > 0
		rsign := rsample > 0

		if i > 0 && lsign != llastSign {
			lzcr++
		}

		if i > 0 && rsign != rlastSign {
			rzcr++
		}

		llastSign = lsign
		rlastSign = rsign
	}

	lrms := math.Sqrt(lsum / float64(len(s)/2))
	rrms := math.Sqrt(rsum / float64(len(s)/2))

	if lrms > 100 && lzcr > 5 {
		ae.leftChannel = true
	} else if rrms > 100 && rzcr > 5 {
		ae.leftChannel = false
	}

	switch ae.leftChannel {
	case true:
		for i := 0; i < frames && i < len(m); i++ {
			m[i] = s[i*2]
		}
	case false:
		for i := 0; i < frames && i < len(m); i++ {
			m[i] = s[i*2+1]
		}
	}

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

func getRms(buffer []int16) float64 {
	var sum float64

	for _, sample := range buffer {
		val := float64(sample)
		sum += val * val
	}

	return math.Sqrt(sum / float64(len(buffer)))
}

func getRmsAndZcr(buffer []int16) (rms float64, zcr int) {
	var (
		sum      float64
		lastSign bool
	)

	for i, sample := range buffer {
		val := float64(sample)
		sum += val * val

		sign := sample > 0
		if i > 0 && sign != lastSign {
			zcr++
		}
		lastSign = sign
	}

	rms = math.Sqrt(sum / float64(len(buffer)))
	return rms, zcr
}

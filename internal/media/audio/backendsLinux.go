//go:build linux

package audio

import "github.com/gen2brain/malgo"

var audioBackends = []malgo.Backend{
	malgo.BackendPulseaudio,
	malgo.BackendAlsa,
}

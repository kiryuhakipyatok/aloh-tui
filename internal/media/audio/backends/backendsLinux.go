//go:build linux

package backends
import "github.com/gen2brain/malgo"

var AudioBackends = []malgo.Backend{
	malgo.BackendPulseaudio,
	malgo.BackendAlsa,
}

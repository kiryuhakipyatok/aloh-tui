//go:build windows

package audio

import "github.com/gen2brain/malgo"

var audioBackends = []malgo.Backend{
	malgo.BackendWasapi,
}

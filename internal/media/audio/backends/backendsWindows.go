//go:build windows

package backends

import "github.com/gen2brain/malgo"

var AudioBackends = []malgo.Backend{
	malgo.BackendWasapi,
}

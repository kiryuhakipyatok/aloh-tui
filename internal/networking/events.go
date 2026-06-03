package networking

import (
	"time"

	alohnetwork "github.com/kiryuhakipyatok/aloh-networking"
)

func MuteMicEvent(state bool) alohnetwork.Event {
	t := time.Now().UTC().Unix()
	return alohnetwork.Event{
		Typee:     alohnetwork.MIC_MUTE,
		State:     state,
		Timestamp: t,
	}
}

func MuteFullEvent(state bool) alohnetwork.Event {
	t := time.Now().UTC().Unix()
	return alohnetwork.Event{
		Typee:     alohnetwork.FULL_MUTE,
		State:     state,
		Timestamp: t,
	}
}
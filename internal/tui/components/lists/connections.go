package lists

import (
	"fmt"
)

type ConnectionItem struct {
	Nickname          string
	VolumeCoefficient float32
	Muted             bool
}

func (ci ConnectionItem) Title() string {
	return ci.Nickname
}
func (ci ConnectionItem) Description() string {
	return fmt.Sprintf("volume: %.1f, muted: %t", ci.VolumeCoefficient, ci.Muted)
}
func (ci ConnectionItem) FilterValue() string {
	return ci.Nickname
}

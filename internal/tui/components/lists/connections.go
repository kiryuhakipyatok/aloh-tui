package lists

import (
	"fmt"
	"strings"
)

type ConnectionItem struct {
	Nickname          string
	VolumeCoefficient float32
	Muted             bool
	BFTag                string
}

func (ci ConnectionItem) Title() string {
	name := ci.Nickname
	if ci.BFTag != "" {
		name = ci.BFTag + " " + name
	}
	return name
}
func (ci ConnectionItem) Description() string {
	return strings.TrimSpace(fmt.Sprintf("volume: %.1f, muted: %t", ci.VolumeCoefficient, ci.Muted))
}
func (ci ConnectionItem) FilterValue() string {
	return ci.Nickname
}

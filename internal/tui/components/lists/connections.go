package lists

import (
	"fmt"
)

type ConnectionItem struct {
	Nickname          string
	VolumeCoefficient float32
	Muted             bool
	Relation             string
}

func (ci ConnectionItem) Title() string {
	name := ci.Nickname
	if ci.Relation != "" {
		name = ci.Relation + " " + name
	}
	return name
}
func (ci ConnectionItem) Description() string {
	return ci.DynamicDescription(false)
}
func (ci ConnectionItem) FilterValue() string {
	return ci.Nickname
}

func (ci ConnectionItem) DynamicDescription(isSelected bool) string {
	if isSelected {
		return fmt.Sprintf("volume: %.1f, muted: %t, ALT+UP/DN to set volume, ALT+F to switch mute", ci.VolumeCoefficient, ci.Muted)
	}

	return fmt.Sprintf("volume: %.1f, muted: %t", ci.VolumeCoefficient, ci.Muted)
}

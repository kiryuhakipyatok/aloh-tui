package lists

import "fmt"

type MicItem struct {
	Name       string
	Channels   uint32
	SampleRate uint32
	Format     string
	Current    string
}

func (mi MicItem) Title() string {
	return mi.Name
}
func (mi MicItem) Description() string {
	return fmt.Sprintf("channels: %d, sample rate: %d, format: %s %s", mi.Channels, mi.SampleRate, mi.Format, mi.Current)
}
func (mi MicItem) FilterValue() string {
	return mi.Name
}

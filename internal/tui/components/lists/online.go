package lists

import (
	"fmt"
	"strings"
)

type OnlineItem struct {
	Name        string
	Connections []string
	BFTag          string
}

func (oi OnlineItem) Title() string {
	name := oi.Name
	if oi.BFTag != "" {
		name = oi.BFTag + " " + name
	}
	return name
}
func (oi OnlineItem) Description() string {
	desc := "alone"
	if len(oi.Connections) > 0 {
		desc = strings.TrimSpace(fmt.Sprintf("with: %s", strings.Join(oi.Connections, "  ·  ")))
	}
	return desc
}
func (oi OnlineItem) FilterValue() string {
	return oi.Name
}
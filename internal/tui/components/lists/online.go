package lists

import (
	"fmt"
	"strings"
)

type OnlineItem struct {
	Name        string
	Connections []string
	BFTag       string
}

func (oi OnlineItem) Title() string {
	name := oi.Name
	if oi.BFTag != "" {
		name = oi.BFTag + " " + name
	}
	return name
}

func (oi OnlineItem) Description() string {
	return oi.DynamicDescription(false)
}

func (oi OnlineItem) FilterValue() string {
	return oi.Name
}

func (oi OnlineItem) DynamicDescription(isSelected bool) string {
	ent := ""
	if isSelected {
		ent = ", ENTER to connect"
	}
	desc := "alone" + ent
	if len(oi.Connections) > 0 {
		conns := strings.Join(oi.Connections, "  ·  ")
		desc = fmt.Sprintf("with: %s%s", conns, ent)
	}
	return desc
}

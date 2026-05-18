package lists

import (
	"fmt"
	"strings"
)

type OnlineItem struct {
	Name        string
	Connections []string
}

func (oi OnlineItem) Title() string {
	return oi.Name
}
func (oi OnlineItem) Description() string {
	desc := "alone"
	if len(oi.Connections) > 0 {
		desc = fmt.Sprintf("with: %s", strings.Join(oi.Connections, "  ·  "))
	}
	return desc
}
func (oi OnlineItem) FilterValue() string {
	return oi.Name
}
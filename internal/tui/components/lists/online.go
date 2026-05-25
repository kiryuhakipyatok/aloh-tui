package lists

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
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
	desc := "alone"
	conns := oi.Connections
	if len(conns) > 0 {
		desc = strings.TrimSpace(fmt.Sprintf("with: %s", strings.Join(conns, "  ·  ")))
	}
	return desc
}

func (oi OnlineItem) FilterValue() string {
	return oi.Name
}

type DynamicOnlineDelegate struct {
	list.DefaultDelegate
}

type dynamicOnlineItem struct {
	isSelected bool
	OnlineItem
}

func (doi dynamicOnlineItem) Description() string {
	if doi.isSelected {
		ent := ", ENTER to connect"
		desc := "alone" + ent
		conns := doi.Connections
		if len(conns) > 0 {
			desc = strings.TrimSpace(fmt.Sprintf("with: %s%s", strings.Join(conns, "  ·  "), ent))
		}
		return desc
	}

	return doi.OnlineItem.Description()
}

func (dd DynamicOnlineDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	oi, ok := item.(OnlineItem)
	if !ok {
		dd.DefaultDelegate.Render(w, m, index, item)
		return
	}

	isSelected := m.Index() == index

	dynamicOnlineItem := dynamicOnlineItem{
		isSelected: isSelected,
		OnlineItem: oi,
	}

	dd.DefaultDelegate.Render(w, m, index, dynamicOnlineItem)
}

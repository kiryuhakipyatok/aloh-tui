package lists

import (
	"io"

	"github.com/charmbracelet/bubbles/list"
)

type DynamicItem interface {
	Title() string
	Description() string
	list.Item
	DynamicDescription(isSelected bool) string
}

type dynamicItem struct {
	DynamicItem
	isSelected bool
}

func (di dynamicItem) Description() string {
	return di.DynamicItem.DynamicDescription(di.isSelected)
}

type DynamicDelegate struct {
	list.DefaultDelegate
}

func (dd DynamicDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {

	if i, ok := item.(DynamicItem); ok {
		isSelected := m.Index() == index

		dynamicItem := dynamicItem{
			isSelected:  isSelected,
			DynamicItem: i,
		}

		dd.DefaultDelegate.Render(w, m, index, dynamicItem)
		return
	}

	dd.DefaultDelegate.Render(w, m, index, item)
}

package lists

import (
	"fmt"
)

type ApearenceItem struct {
	Id      uint
	Name    string
	Desc    string
	Enabled bool
}

func (ai ApearenceItem) Title() string {
	return ai.Name
}
func (ai ApearenceItem) Description() string {
	return ai.DynamicDescription(false)
}
func (ai ApearenceItem) FilterValue() string {
	return ai.Name
}

func (ai ApearenceItem) DynamicDescription(isSelected bool) string {
	if isSelected {
		return fmt.Sprintf("%s, state: %t, ENTER to switch", ai.Desc, ai.Enabled)
	}

	return fmt.Sprintf("%s, state: %t", ai.Desc, ai.Enabled)
}

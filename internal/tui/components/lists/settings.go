package lists

import (
	"fmt"
)

type SettingsItem struct {
	Id      uint
	Name    string
	Desc    string
	Enabled bool
}

func (si SettingsItem) Title() string {
	return si.Name
}
func (si SettingsItem) Description() string {
	return si.DynamicDescription(false)
}
func (si SettingsItem) FilterValue() string {
	return si.Name
}

func (si SettingsItem) DynamicDescription(isSelected bool) string {
	if isSelected {
		return fmt.Sprintf("%s, state: %t, ENTER to switch", si.Desc, si.Enabled)
	}

	return fmt.Sprintf("%s, state: %t", si.Desc, si.Enabled)
}

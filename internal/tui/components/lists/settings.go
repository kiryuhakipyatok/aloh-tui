package lists

import "fmt"

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
	return fmt.Sprintf("%s, state: %t", si.Desc, si.Enabled)
}
func (si SettingsItem) FilterValue() string {
	return si.Name
}

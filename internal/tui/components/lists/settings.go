package lists

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

type SettingsItem struct {
	Id   uint
	Name string
	Desc string
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
		return fmt.Sprintf("%s, ENTER to fall", si.Desc)
	}

	return si.Desc
}

const (
	DEVICES_SETTINGS = iota
	BINDS_SETTINGS
	AUDIO_SETTINGS
	NOTIFICATIONS_SETTINGS
	ACCOUNT_SETTINGS
	NICKNAME_SETTINGS
	COLOR_SETTINGS
	TAGLINE_SETTINGS
	PASSWORD_SETTINGS
	HEADPHONES_SETTINGS
	WEBCAM_SETTINGS
	MICROPHONE_SETTINGS
)

func SetupSettingsList(ls ListSetup) SettingsList {

	settings := []list.Item{
		SettingsItem{
			Id:   DEVICES_SETTINGS,
			Name: "devices",
			Desc: "pick microphone or headphones",
		},
		SettingsItem{
			Id:   BINDS_SETTINGS,
			Name: "binds",
			Desc: "manage binds",
		},
		SettingsItem{
			Id:   AUDIO_SETTINGS,
			Name: "audio",
			Desc: "manage audio settings",
		},
		SettingsItem{
			Id:   NOTIFICATIONS_SETTINGS,
			Name: "notifications",
			Desc: "manage notifications",
		},
		SettingsItem{
			Id:   ACCOUNT_SETTINGS,
			Name: "account",
			Desc: "manage your account",
		},
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(ls.ThemeColor)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(ls.SubColor)
	delegate.Styles.NormalTitle = delegate.Styles.NormalDesc.Foreground(ls.NormalTitleColor)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(ls.NormalDescColor)
	delegate.SetSpacing(1)

	dd := DynamicDelegate{DefaultDelegate: delegate}

	l := list.New(settings, dd, 30, 30)

	l.Select(0)
	l.DisableQuitKeybindings()
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.SetFilteringEnabled(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)

	return SettingsList{
		DefList: DefList{
			LipList:     l,
			LipDelegate: dd,
			themeColor:  ls.ThemeColor,
			subColor:    ls.SubColor,
		},
	}
}

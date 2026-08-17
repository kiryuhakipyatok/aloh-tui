package lists

import (
	"aloh-tui/internal/entities/users"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

type DeviceTypeItem struct {
	Id   uint
	Name string
	Desc string
}

func (dti DeviceTypeItem) Title() string {
	return dti.Name
}
func (dti DeviceTypeItem) Description() string {
	return dti.DynamicDescription(false)
}
func (dti DeviceTypeItem) FilterValue() string {
	return dti.Name
}

func (dti DeviceTypeItem) DynamicDescription(isSelected bool) string {
	if isSelected {
		return fmt.Sprintf("%s, ENTER to fall", dti.Desc)
	}

	return dti.Desc
}

func SetupDevicesTypeList(user *users.User, ls ListSetup) SettingsList {

	devicesTypes := []list.Item{
		DeviceTypeItem{
			Id:   HEADPHONES_SETTINGS,
			Name: "headphones",
			Desc: "change your headphones",
		},
		DeviceTypeItem{
			Id:   MICROPHONE_SETTINGS,
			Name: "microphone",
			Desc: "change your microphone",
		},
		DeviceTypeItem{
			Id:   WEBCAM_SETTINGS,
			Name: "webcam",
			Desc: "change your webcam",
		},
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(ls.ThemeColor)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(ls.SubColor)
	delegate.Styles.NormalTitle = delegate.Styles.NormalDesc.Foreground(ls.NormalTitleColor)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(ls.NormalDescColor)
	delegate.SetSpacing(1)

	dd := DynamicDelegate{DefaultDelegate: delegate}

	l := list.New(devicesTypes, dd, 30, 30)

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

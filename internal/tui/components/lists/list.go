package lists

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

type DeviceList struct {
	DefList
}

type SwitcherList struct {
	DefList
}

type OnlineList struct {
	DefList
}

type ConnectionsList struct {
	DefList
}

type FriendsReqsList struct {
	DefList
}

type SettingsList struct {
	DefList
}

type DefList struct {
	LipList     list.Model
	LipDelegate DynamicDelegate
	themeColor  lipgloss.Color
	subColor    lipgloss.Color
}

type ListSetup struct {
	ThemeColor       lipgloss.Color
	SubColor         lipgloss.Color
	NormalDescColor  lipgloss.Color
	NormalTitleColor lipgloss.AdaptiveColor
}

func SetListVisible(l *DefList, vis bool) {
	if vis {
		l.LipDelegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(l.themeColor)
		l.LipDelegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(l.subColor)
	} else {
		l.LipDelegate.Styles.SelectedTitle = l.LipDelegate.Styles.NormalTitle
		l.LipDelegate.Styles.SelectedDesc = l.LipDelegate.Styles.NormalDesc
	}

	l.LipList.SetDelegate(l.LipDelegate)
}

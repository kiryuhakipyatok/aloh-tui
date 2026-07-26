package lists

import (
	"aloh-tui/internal/entities/users"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

type AccountItem struct {
	Id   uint
	Name string
	Desc string
}

func (ai AccountItem) Title() string {
	return ai.Name
}
func (ai AccountItem) Description() string {
	return ai.DynamicDescription(false)
}
func (ai AccountItem) FilterValue() string {
	return ai.Name
}

func (ai AccountItem) DynamicDescription(isSelected bool) string {
	if isSelected {
		return fmt.Sprintf("%s, ENTER to fall", ai.Desc)
	}

	return ai.Desc
}

func SetupAccountList(user *users.User, ls ListSetup) SettingsList {

	accounts := []list.Item{
		AccountItem{
			Id:   NICKNAME_SETTINGS,
			Name: "nickname",
			Desc: "change your nickname",
		},
		AccountItem{
			Id:   COLOR_SETTINGS,
			Name: "color",
			Desc: "change your color",
		},
		AccountItem{
			Id:   TAGLINE_SETTINGS,
			Name: "tagline",
			Desc: "change your tagline",
		},
		AccountItem{
			Id:   PASSWORD_SETTINGS,
			Name: "password",
			Desc: "change your password",
		},
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(ls.ThemeColor)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(ls.SubColor)
	delegate.Styles.NormalTitle = delegate.Styles.NormalDesc.Foreground(ls.NormalTitleColor)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(ls.NormalDescColor)
	delegate.SetSpacing(1)

	dd := DynamicDelegate{DefaultDelegate: delegate}

	l := list.New(accounts, dd, 30, 30)

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

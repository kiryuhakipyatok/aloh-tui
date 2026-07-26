package lists

import (
	"aloh-tui/internal/entities/users"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	AUDIO = iota
	NOTIFICATIONS
	APEREANCE
)

const (
	HARD_DENOISE = iota
	SOFT_DENOISE
	AEC
	EQUALIZER
	AUDIO_N
	DESKTOP_N
	APP_N
	TIME
	DATE
	ZONE
)

type SwitcherItem struct {
	Id      uint
	Name    string
	Desc    string
	Enabled bool
}

func (si SwitcherItem) Title() string {
	return si.Name
}
func (si SwitcherItem) Description() string {
	return si.DynamicDescription(false)
}
func (si SwitcherItem) FilterValue() string {
	return si.Name
}

func (si SwitcherItem) DynamicDescription(isSelected bool) string {
	if isSelected {
		return fmt.Sprintf("%s, state: %t, ENTER to switch", si.Desc, si.Enabled)
	}

	return fmt.Sprintf("%s, state: %t", si.Desc, si.Enabled)
}

func SetupSwitcherList(user *users.User, switchersType uint, ls ListSetup) SwitcherList {

	var switchers []list.Item

	denoises := user.GetDenoises()

	switch switchersType {
	case AUDIO:
		switchers = []list.Item{
			SwitcherItem{
				Id:      HARD_DENOISE,
				Name:    "hard denoise",
				Desc:    "reduce noise hard",
				Enabled: denoises.HardDenoise,
			},
			SwitcherItem{
				Id:      SOFT_DENOISE,
				Name:    "soft denoise",
				Desc:    "reduce noise soft",
				Enabled: denoises.SoftDenoise,
			},
			SwitcherItem{
				Id:      AEC,
				Name:    "echocanceller",
				Desc:    "reduce echo",
				Enabled: user.Data.Setup.Audio.AEC,
			},
			SwitcherItem{
				Id:      EQUALIZER,
				Name:    "equalizer",
				Desc:    "reduce low freqs and increase high",
				Enabled: user.Data.Setup.Audio.Filter,
			},
		}

	case APEREANCE:
		switchers = []list.Item{
			SwitcherItem{
				Id:      TIME,
				Name:    "show time",
				Desc:    "current time in top-right corner",
				Enabled: user.Data.Setup.Appereance.ShowTime,
			},
			SwitcherItem{
				Id:      DATE,
				Name:    "show date",
				Desc:    "current date in top-right corner",
				Enabled: user.Data.Setup.Appereance.ShowDate,
			},
			SwitcherItem{
				Id:      ZONE,
				Name:    "show zone",
				Desc:    "current zone in top-right corner",
				Enabled: user.Data.Setup.Appereance.ShowZone,
			},
		}

	case NOTIFICATIONS:
		switchers = []list.Item{
			SwitcherItem{
				Id:      APP_N,
				Name:    "app notifications",
				Desc:    "notifications in app intreface",
				Enabled: user.Data.Setup.Notifications.AppNotifications,
			},
			SwitcherItem{
				Id:      AUDIO_N,
				Name:    "audio notifications",
				Desc:    "notifications with sound",
				Enabled: user.Data.Setup.Notifications.AudioNotifications,
			},
			SwitcherItem{
				Id:      DESKTOP_N,
				Name:    "desktop notifications",
				Desc:    "notifications with desktop notifying",
				Enabled: user.Data.Setup.Notifications.DesktopNotifications,
			},
		}

	default:
		return SwitcherList{}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(ls.ThemeColor)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(ls.SubColor)

	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(ls.NormalDescColor)
	delegate.Styles.NormalTitle = delegate.Styles.NormalDesc.Foreground(ls.NormalTitleColor)
	delegate.SetSpacing(1)

	dd := DynamicDelegate{DefaultDelegate: delegate}

	l := list.New(switchers, dd, 30, 30)

	l.Select(0)
	l.DisableQuitKeybindings()
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.SetFilteringEnabled(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)

	return SwitcherList{
		DefList: DefList{
			LipList:     l,
			LipDelegate: dd,
			themeColor:  ls.ThemeColor,
			subColor:    ls.SubColor,
		},
	}
}

func (l *SwitcherList) UpdateSwitcherItemList(switcherId uint) tea.Cmd {
	var cmd tea.Cmd
	items := l.LipList.Items()
	for i, v := range items {
		s, ok := v.(SwitcherItem)
		if !ok {
			continue
		}
		if s.Id == switcherId {
			s.Enabled = !s.Enabled
			cmd = l.LipList.SetItem(i, s)
			return cmd
		}
	}
	return cmd
}

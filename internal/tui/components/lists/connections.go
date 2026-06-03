package lists

import (
	"aloh-tui/internal/entities/users"
	"fmt"
	"slices"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type ConnectionItem struct {
	Nickname          string
	VolumeCoefficient float32
	Muted             bool
	Relation          string
}

func (ci ConnectionItem) Title() string {
	name := ci.Nickname
	if ci.Relation != "" {
		name = ci.Relation + " " + name
	}
	return name
}
func (ci ConnectionItem) Description() string {
	return ci.DynamicDescription(false)
}
func (ci ConnectionItem) FilterValue() string {
	return ci.Nickname
}

func (ci ConnectionItem) DynamicDescription(isSelected bool) string {
	if isSelected {
		return fmt.Sprintf("volume: %.1f, muted: %t, ALT+UP/DN to set volume, ALT+F to switch mute", ci.VolumeCoefficient, ci.Muted)
	}

	return fmt.Sprintf("volume: %.1f, muted: %t", ci.VolumeCoefficient, ci.Muted)
}

func SetupConnestionsList(ls ListSetup) ConnectionsList {
	conns := make([]list.Item, 0)

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(ls.ThemeColor)
	delegate.Styles.SelectedTitle = lipgloss.NewStyle()
	delegate.Styles.NormalTitle = delegate.Styles.NormalDesc.Foreground(ls.NormalTitleColor)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(ls.NormalDescColor)
	delegate.ShowDescription = true
	delegate.SetSpacing(1)

	dd := DynamicDelegate{DefaultDelegate: delegate}

	l := list.New(conns, dd, 30, 30)
	l.DisableQuitKeybindings()
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.SetFilteringEnabled(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)

	return ConnectionsList{
		DefList: DefList{
			LipList:     l,
			LipDelegate: dd,
			themeColor:  ls.ThemeColor,
			subColor:    ls.SubColor,
		},
	}
}

func (l *ConnectionsList) UpdateConnectionItemList(nickname string, volume float32, muted bool) tea.Cmd {
	var cmds []tea.Cmd
	items := l.LipList.Items()
	for i, v := range items {
		conn, ok := v.(ConnectionItem)
		if !ok {
			continue
		}
		if conn.Nickname == nickname {
			conn.VolumeCoefficient = volume
			conn.Muted = muted
		}

		cmds = append(cmds, l.LipList.SetItem(i, conn))
	}
	return tea.Batch(cmds...)
}

func (l *ConnectionsList) UpdateConnectionsList(user *users.User, connections []string) tea.Cmd {
	names := make([]string, 0, len(connections))
	var clearName string
	for _, name := range connections {
		clearName = ansi.Strip(name)
		if !user.IsBlocked(clearName) {
			names = append(names, name)
		}
	}

	slices.Sort(names)

	newItems := make([]list.Item, 0, len(names))
	for _, name := range names {
		var vc float32 = 1
		var muted bool
		clearName = ansi.Strip(name)
		us, ok := user.Data.Setup.Audio.UsersSetup[clearName]
		if ok {
			vc = us.VolumeCoefficient
			muted = us.Muted
		}
		var rel string
		if user.Data.Statistics.BestFriend.Nickname == clearName {
			rel = user.Data.Setup.Appereance.BestFriendTag
		}
		newItems = append(newItems, ConnectionItem{
			Nickname:          name,
			VolumeCoefficient: vc,
			Muted:             muted,
			Relation:          rel,
		})
	}

	return l.LipList.SetItems(newItems)
}

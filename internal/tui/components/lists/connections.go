package lists

import (
	"aloh-tui/internal/entities/users"
	"aloh-tui/internal/tui/components/styles"
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/kiryuhakipyatok/aloh-signalling/pkg/errs"
)

type ConnectionItem struct {
	Identity            users.Identity
	VolumeCoefficient   float32
	Muted               bool
	PersonalHardDenoise bool
	PersonalSoftDenoise bool
	Relation            string
}

func (ci ConnectionItem) Title() string {
	name := ci.Identity.Nickname
	if ci.Relation != "" {
		name = ci.Relation + " " + name
	}
	return name
}
func (ci ConnectionItem) Description() string {
	return ci.DynamicDescription(false)
}
func (ci ConnectionItem) FilterValue() string {
	return ci.Identity.Nickname
}

func (ci ConnectionItem) DynamicDescription(isSelected bool) string {
	if isSelected {
		return fmt.Sprintf("volume: %.1f, muted: %t, psd: %t, phd: %t, SELECTED",
			ci.VolumeCoefficient, ci.Muted, ci.PersonalSoftDenoise, ci.PersonalHardDenoise)
	}

	return fmt.Sprintf("volume: %.1f, muted: %t, psd: %t, phd: %t",
		ci.VolumeCoefficient, ci.Muted, ci.PersonalSoftDenoise, ci.PersonalHardDenoise)
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

func (l *ConnectionsList) UpdateConnectionItemList(iden users.Identity, volume float32, muted, phd, psd bool) tea.Cmd {
	var cmds []tea.Cmd
	items := l.LipList.Items()
	for i, v := range items {
		conn, ok := v.(ConnectionItem)
		if !ok {
			continue
		}
		if conn.Identity == iden {
			conn.VolumeCoefficient = volume
			conn.Muted = muted
			conn.PersonalHardDenoise = phd
			conn.PersonalSoftDenoise = psd
		}

		cmds = append(cmds, l.LipList.SetItem(i, conn))
	}
	return tea.Batch(cmds...)
}

func (l *ConnectionsList) UpdateConnectionsList(user *users.User, connections []users.Identity,
	colors map[uuid.UUID]styles.UserColors) tea.Cmd {
	slices.SortFunc(connections, func(c1, c2 users.Identity) int {
		return strings.Compare(c1.Nickname, c2.Nickname)
	})

	newItems := make([]list.Item, 0, len(connections))
	for _, conn := range connections {
		var (
			vc    float32 = 1
			muted bool
			name  = conn.Nickname
		)
		us, ok := user.Data.Setup.Audio.UsersSetup[name]
		if ok {
			vc = us.VolumeCoefficient
			muted = us.Muted
		}
		var rel string
		if user.Data.Statistics.BestFriend.Identity.Nickname == name {
			rel = user.Data.Setup.Appereance.BestFriendTag
		}
		coloredName := lipgloss.NewStyle().Foreground(colors[conn.ID].MainColor).Render(name)
		conn.Nickname = coloredName
		newItems = append(newItems, ConnectionItem{
			Identity:          conn,
			VolumeCoefficient: vc,
			Muted:             muted,
			Relation:          rel,
		})
	}

	return l.LipList.SetItems(newItems)
}

func (l *ConnectionsList) GetConnectionItem(iden users.Identity) (ConnectionItem, error) {
	var ci ConnectionItem

	items := l.LipList.Items()
	for _, v := range items {
		conn, ok := v.(ConnectionItem)
		if !ok {
			continue
		}
		if conn.Identity == iden {
			ci = conn
			break
		}
	}
	if ci.Identity.ID == uuid.Nil {
		return ci, errs.ErrNotFoundBase
	}
	return ci, errs.ErrNotFoundBase
}

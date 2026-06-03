package lists

import (
	"aloh-tui/internal/entities/users"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type OnlineItem struct {
	Name        string
	Connections []string
	BFTag       string
}

func (oi OnlineItem) Title() string {
	name := oi.Name
	if oi.BFTag != "" {
		name = oi.BFTag + " " + name
	}
	return name
}

func (oi OnlineItem) Description() string {
	return oi.DynamicDescription(false)
}

func (oi OnlineItem) FilterValue() string {
	return oi.Name
}

func (oi OnlineItem) DynamicDescription(isSelected bool) string {
	ent := ""
	if isSelected {
		ent = ", ENTER to connect"
	}
	desc := "alone" + ent
	if len(oi.Connections) > 0 {
		conns := strings.Join(oi.Connections, " · ")
		desc = fmt.Sprintf("with: %s%s", conns, ent)
	}
	return desc
}

func SetupOnlineList(curOnline map[string][]string, ls ListSetup) OnlineList {
	online := make([]list.Item, len(curOnline))

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(ls.ThemeColor)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(ls.SubColor)
	delegate.Styles.NormalTitle = delegate.Styles.NormalDesc.Foreground(ls.NormalTitleColor)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(ls.NormalDescColor)
	delegate.SetSpacing(1)

	dd := DynamicDelegate{DefaultDelegate: delegate}

	l := list.New(online, dd, 30, 30)
	l.DisableQuitKeybindings()
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.SetFilteringEnabled(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)

	return OnlineList{
		DefList: DefList{
			LipList:     l,
			LipDelegate: dd,
			themeColor:  ls.ThemeColor,
			subColor:    ls.SubColor,
		},
	}
}

func (l *OnlineList) UpdateOnlineList(user *users.User, curOnline map[string][]string) tea.Cmd {
	names := make([]string, 0, len(curOnline))
	bfTag := user.Data.Setup.Appereance.BestFriendTag
	bfNick := user.Data.Statistics.BestFriend.Nickname
	banTag := user.Data.Setup.Appereance.BanTag
	var wg sync.WaitGroup
	for name := range curOnline {
		names = append(names, name)
		wg.Go(func() {
			for i, c := range curOnline[name] {
				if bfNick == c {
					curOnline[name][i] = bfTag + " " + c
				} else if user.IsBlocked(c) {
					curOnline[name][i] = banTag + " " + c
				}
			}
		})
	}
	wg.Wait()
	slices.Sort(names)

	newItems := make([]list.Item, 0, len(names))

	for _, name := range names {
		var rel string
		if bfNick == name {
			rel = bfTag
		}
		newItems = append(newItems, OnlineItem{
			Name:        name,
			Connections: curOnline[name],
			BFTag:       rel,
		})
	}

	return l.LipList.SetItems(newItems)
}

package lists

import (
	"aloh-tui/internal/entities/users"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

type FriendItem struct {
	Identity         users.Identity
	Tagline          string
	Relation         string
	Connections      []users.Identity
	ConnectionsNicks []string
	IsOnline         bool
}

func (fi FriendItem) Title() string {
	return fi.Relation + " " + fi.Identity.Nickname
}

func (fi FriendItem) Description() string {
	return fi.DynamicDescription(false)
}

func (fi FriendItem) FilterValue() string {
	return fi.Identity.Nickname
}

func (fi FriendItem) DynamicDescription(isSelected bool) string {
	if !fi.IsOnline {
		ofStr := "offline"
		if fi.Tagline != "" {
			ofStr = strings.TrimSpace(fi.Tagline + " | " + ofStr)
		}
		return ofStr
	}
	ent := ""
	if isSelected {
		ent = ", ENTER to connect"
	}
	onStr := "alone" + ent
	if fi.Tagline != "" {
		onStr = strings.TrimSpace(fi.Tagline + " | " + onStr)
	}
	if len(fi.ConnectionsNicks) > 0 {
		conns := strings.Join(fi.ConnectionsNicks, " · ")
		onStr = strings.TrimSpace(fi.Tagline + fmt.Sprintf(" | with: %s%s", conns, ent))
	}
	return onStr
}

func SetupFriendsList(user *users.User, ls ListSetup) FriendsList {
	bfTag := user.Data.Setup.Appereance.BestFriendTag
	bfIden := user.Data.Statistics.BestFriend.Identity
	friends := user.Data.Personal.Friends
	friendsItems := make([]list.Item, 0, len(friends))

	for _, f := range friends {
		var rel string
		if bfIden == f.Identity {
			rel = bfTag
		}

		friendsItems = append(friendsItems, FriendItem{
			Identity:         f.Identity,
			Tagline:          f.Tagline,
			Connections:      nil,
			ConnectionsNicks: nil,
			Relation:         rel,
			IsOnline:         false,
		})
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(ls.ThemeColor)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(ls.SubColor)
	delegate.Styles.NormalTitle = delegate.Styles.NormalDesc.Foreground(ls.NormalTitleColor)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(ls.NormalDescColor)
	delegate.SetSpacing(1)

	dd := DynamicDelegate{DefaultDelegate: delegate}

	l := list.New(friendsItems, dd, 30, 30)
	l.DisableQuitKeybindings()
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.SetFilteringEnabled(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)

	return FriendsList{
		DefList: DefList{
			LipList:     l,
			LipDelegate: dd,
			themeColor:  ls.ThemeColor,
			subColor:    ls.SubColor,
		},
	}
}

func (l *FriendsList) UpdateFriendsList(user *users.User, curOnline map[uuid.UUID][]users.Identity) tea.Cmd {
	friends := user.GetFriends()
	bfTag := user.Data.Setup.Appereance.BestFriendTag
	bfIden := user.Data.Statistics.BestFriend.Identity
	banTag := user.Data.Setup.Appereance.BanTag

	//slices.Sort(friends)

	curOLen := len(curOnline)
	frLen := len(friends)

	frItems := make([]list.Item, 0, frLen)

	offlineItems := make([]list.Item, 0, frLen-curOLen)
	onlineItems := make([]list.Item, 0, curOLen)

	for _, f := range friends {
		id := f.ID
		var (
			nicknames []string
			isOnline  bool
			conns     []users.Identity
		)
		if cons, ok := curOnline[id]; ok {
			nicks := make([]string, 0, len(cons))
			for _, c := range cons {
				n := c.Nickname
				if bfIden.ID == c.ID {
					n = bfTag + " " + c.Nickname
				} else if user.IsBlocked(c) {
					n = banTag + " " + c.Nickname
				}
				nicks = append(nicks, n)
			}
			nicknames = nicks
			isOnline = true
			conns = curOnline[id]
		}

		var rel string
		if bfIden == f.Identity {
			rel = bfTag
		}

		fi := FriendItem{
			Identity:         f.Identity,
			Tagline:          f.Tagline,
			Connections:      conns,
			ConnectionsNicks: nicknames,
			Relation:         rel,
			IsOnline:         isOnline,
		}

		if fi.IsOnline {
			onlineItems = append(onlineItems, fi)
		} else {
			offlineItems = append(offlineItems, fi)
		}
	}

	frItems = append(onlineItems, offlineItems...)

	return l.LipList.SetItems(frItems)
}

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
	Id          uuid.UUID
	Name        string
	Tagline     string
	Connections []string
	BFTag       string
	IsOnline    bool
}

func (fi FriendItem) Title() string {
	name := fi.Name
	if fi.BFTag != "" {
		name = fi.BFTag + " " + name
	}
	return name
}

func (fi FriendItem) Description() string {
	return fi.DynamicDescription(false)
}

func (fi FriendItem) FilterValue() string {
	return fi.Name
}

func (fi FriendItem) DynamicDescription(isSelected bool) string {
	if !fi.IsOnline {
		return "offline"
	}
	ent := ""
	if isSelected {
		ent = ", ENTER to connect"
	}
	desc := fi.Tagline + " | alone" + ent
	if len(fi.Connections) > 0 {
		conns := strings.Join(fi.Connections, " · ")
		desc = fmt.Sprintf("with: %s%s", conns, ent)
	}
	return desc
}

func SetupFriendsList(user *users.User, ls ListSetup) FriendsList {
	bfTag := user.Data.Setup.Appereance.BestFriendTag
	bfNick := user.Data.Statistics.BestFriend.Nickname
	friends := user.Data.Personal.Friends
	friendsItems := make([]list.Item, 0, len(friends))

	for _, f := range friends {
		var rel string
		name := f.Nickname
		if bfNick == name {
			rel = bfTag
		}

		friendsItems = append(friendsItems, FriendItem{
			Id:          f.ID,
			Tagline:     f.Tagline,
			Name:        name,
			Connections: nil,
			BFTag:       rel,
			IsOnline:    false,
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

func (l *FriendsList) UpdateFriendsList(user *users.User, curOnline map[uuid.UUID][]string) tea.Cmd {
	friends := user.GetFriends()
	bfTag := user.Data.Setup.Appereance.BestFriendTag
	bfNick := user.Data.Statistics.BestFriend.Nickname
	banTag := user.Data.Setup.Appereance.BanTag

	//slices.Sort(friends)

	curOLen := len(curOnline)
	frLen := len(friends)

	frItems := make([]list.Item, 0, frLen)

	offlineItems := make([]list.Item, 0, frLen-curOLen)
	onlineItems := make([]list.Item, 0, curOLen)

	for _, f := range friends {
		id := f.ID
		if _, ok := curOnline[id]; ok {

			for i, c := range curOnline[id] {
				if bfNick == c {
					curOnline[id][i] = bfTag + " " + c
				} else if user.IsBlocked(c) {
					curOnline[id][i] = banTag + " " + c
				}
			}

		}
		f, ok := friends[id]
		if !ok {
			continue
		}
		name := f.Nickname
		var rel string
		if bfNick == name {
			rel = bfTag
		}

		var (
			isOnline bool
			conns    []string
		)

		if _, ok := curOnline[id]; ok {
			isOnline = true
			conns = curOnline[id]
		}

		fi := FriendItem{
			Id:          id,
			Tagline:     f.Tagline,
			Name:        name,
			Connections: conns,
			BFTag:       rel,
			IsOnline:    isOnline,
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

package lists

import (
	"aloh-tui/internal/entities/users"
	"fmt"
	"slices"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FriendReqItem struct {
	Nickname string
	ReqTime  string
}

func (fi FriendReqItem) Title() string {
	return fi.Nickname
}
func (fi FriendReqItem) Description() string {
	return fi.DynamicDescription(false)
}
func (fi FriendReqItem) FilterValue() string {
	return fi.Nickname
}

func (fi FriendReqItem) DynamicDescription(isSelected bool) string {
	if isSelected {
		return fmt.Sprintf("requested in %s, ENTER to accept, ALT+X to deny", fi.ReqTime)
	}

	return fmt.Sprintf("requested in %s", fi.ReqTime)
}

func SetupFriendsReqsList(user *users.User, ls ListSetup) FriendsReqsList {
	fReqs := user.GetFriendsReqs()
	friendsReq := make([]list.Item, 0, len(fReqs))

	for _, f := range fReqs {
		friendsReq = append(friendsReq, FriendReqItem{
			Nickname: f.Nickname,
			ReqTime:  f.ReqTime.Local().Format("2006-01-02"),
		})
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(ls.ThemeColor)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(ls.SubColor)
	delegate.Styles.NormalTitle = delegate.Styles.NormalDesc.Foreground(ls.NormalTitleColor)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(ls.NormalDescColor)
	delegate.SetSpacing(0)

	dd := DynamicDelegate{DefaultDelegate: delegate}

	l := list.New(friendsReq, dd, 30, 30)
	l.Select(0)
	l.DisableQuitKeybindings()
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.SetFilteringEnabled(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)

	return FriendsReqsList{
		DefList: DefList{
			LipList:     l,
			LipDelegate: dd,
			themeColor:  ls.ThemeColor,
			subColor:    ls.SubColor,
		},
	}
}

func (l *FriendsReqsList) UpdateFriendsReqList(user *users.User) tea.Cmd {
	fReqs := user.GetFriendsReqs()
	names := make([]string, 0, len(fReqs))
	for _, f := range fReqs {
		names = append(names, f.Nickname)
	}

	slices.Sort(names)

	newItems := make([]list.Item, 0, len(names))

	for _, name := range names {
		t := time.Now().Format("2006-01-02")
		newItems = append(newItems, FriendReqItem{
			Nickname: name,
			ReqTime:  t,
		})
	}

	return l.LipList.SetItems(newItems)
}

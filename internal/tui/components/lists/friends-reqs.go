package lists

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
)

type FriendReqItem struct {
	Nickname string
	ReqTime  string
}

func (fi FriendReqItem) Title() string {
	return fi.Nickname
}
func (fi FriendReqItem) Description() string {
	return fmt.Sprintf("requested in %s", fi.ReqTime)
}
func (fi FriendReqItem) FilterValue() string {
	return fi.Nickname
}

type DynamicFriendsReqDelegate struct {
	list.DefaultDelegate
}

type dynamicFriendsReqItem struct {
	isSelected bool
	FriendReqItem
}

func (dfi dynamicFriendsReqItem) Description() string {
	if dfi.isSelected {
		return fmt.Sprintf("requested in %s, ENTER to accept, ALT+X to deny", dfi.ReqTime)
	}

	return dfi.FriendReqItem.Description()
}

func (dd DynamicFriendsReqDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	fi, ok := item.(FriendReqItem)
	if !ok {
		dd.DefaultDelegate.Render(w, m, index, item)
		return
	}

	isSelected := m.Index() == index

	dynamicFriendsReqItem := dynamicFriendsReqItem{
		isSelected:    isSelected,
		FriendReqItem: fi,
	}

	dd.DefaultDelegate.Render(w, m, index, dynamicFriendsReqItem)
}

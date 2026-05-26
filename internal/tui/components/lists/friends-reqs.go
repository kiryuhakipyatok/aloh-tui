package lists

import (
	"fmt"
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

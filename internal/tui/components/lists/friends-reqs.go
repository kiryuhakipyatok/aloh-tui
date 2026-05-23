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
	return fmt.Sprintf("requested in %s", fi.ReqTime)
}
func (fi FriendReqItem) FilterValue() string {
	return fi.Nickname
}
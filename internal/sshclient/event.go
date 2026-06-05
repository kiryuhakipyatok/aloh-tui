package sshclient

import "encoding/json"

const (
	NEW_FRIEND_REQ = iota
	ACCEPT_FRIEND
	DENY_FRIEND
	DELETE_FRIEND
	BLOCK_USER
	UNBLOCK_USER
	FRIEND_ONLINE
	FRIEND_OFFLINE
	FRIEND_CONNECTIONS

	SEND_FRIEND_REQ
)

type Event struct {
	Type uint            `json:"type"`
	Data json.RawMessage `json:"data"`
}

type FriendConnsData struct {
	Nickname string   `json:"nickname"`
	Connects []string `json:"connects"`
}

func proccessEvent(event []byte) (Event, error) {
	var e Event
	if err := json.Unmarshal(event, &e); err != nil {
		return e, err
	}
	return e, nil
}

func CastToFriendConnsData(data []byte) (FriendConnsData, error) {
	var fcd FriendConnsData
	if err := json.Unmarshal(data, &fcd); err != nil {
		return fcd, err
	}
	return fcd, nil
}

func CastToNicknameData(data []byte) (string, error) {
	var nickname string
	if err := json.Unmarshal(data, &nickname); err != nil {
		return nickname, err
	}
	return nickname, nil
}

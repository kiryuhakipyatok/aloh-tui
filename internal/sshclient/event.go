package sshclient

import (
	"encoding/json"

	"github.com/google/uuid"
)

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
	UPDATE_HARD_DENOISE
	UPDATE_SOFT_DENOISE
	UPDATE_TAGLINE

	SEND_FRIEND_REQ
)

type Event struct {
	Type uint            `json:"type"`
	Data json.RawMessage `json:"data"`
}

type FriendConnsData struct {
	Id       uuid.UUID `json:"id"`
	Connects []string  `json:"connects"`
}

type FriendPersonal struct {
	ID       uuid.UUID `json:"id"`
	Nickname string    `json:"nickname"`
}

type TaglineData struct {
	Id      uuid.UUID `json:"id"`
	Tagline string    `json:"tagline"`
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

func CastToFriendPersonalData(data []byte) (FriendPersonal, error) {
	var fp FriendPersonal
	if err := json.Unmarshal(data, &fp); err != nil {
		return fp, err
	}
	return fp, nil
}

func CastToNicknameData(data []byte) (string, error) {
	var nickname string
	if err := json.Unmarshal(data, &nickname); err != nil {
		return nickname, err
	}
	return nickname, nil
}

func CastToIdData(data []byte) (uuid.UUID, error) {
	var id uuid.UUID
	if err := json.Unmarshal(data, &id); err != nil {
		return id, err
	}
	return id, nil
}

func CastToTaglineData(data []byte) (TaglineData, error) {
	var tg TaglineData
	if err := json.Unmarshal(data, &tg); err != nil {
		return tg, err
	}
	return tg, nil
}

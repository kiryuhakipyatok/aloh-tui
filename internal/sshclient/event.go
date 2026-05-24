package sshclient

import "encoding/json"

const (
	NEW_FRIEND = iota
	ACCEPT_FRIEND
	DENY_FRIEND
	DELETE_FRIEND
)

type Event struct {
	Type uint   `json:"type"`
	Data string `json:"data"`
}

func proccessEvent(event []byte) (Event, error) {
	var e Event
	if err := json.Unmarshal(event, &e); err != nil {
		return e, err
	}
	return e, nil
}

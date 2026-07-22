package sshclient

import (
	"aloh-tui/pkg/logger"
	"bufio"
	"encoding/json"

	"github.com/google/uuid"
	alohssh "github.com/kiryuhakipyatok/aloh-ssh"
)

const (
	NEW_FRIEND_REQ      = alohssh.NEW_FRIEND_REQ
	ACCEPT_FRIEND       = alohssh.ACCEPT_FRIEND
	DENY_FRIEND         = alohssh.DENY_FRIEND
	DELETE_FRIEND       = alohssh.DELETE_FRIEND
	BLOCK_USER          = alohssh.BLOCK_USER
	UNBLOCK_USER        = alohssh.UNBLOCK_USER
	FRIEND_ONLINE       = alohssh.FRIEND_ONLINE
	FRIEND_OFFLINE      = alohssh.FRIEND_OFFLINE
	FRIEND_CONNECTIONS  = alohssh.FRIEND_CONNECTIONS
	UPDATE_HARD_DENOISE = alohssh.UPDATE_HARD_DENOISE
	UPDATE_SOFT_DENOISE = alohssh.UPDATE_SOFT_DENOISE
	UPDATE_TAGLINE      = alohssh.UPDATE_TAGLINE
	UPDATE_NICKNAME     = alohssh.UPDATE_NICKNAME
	UPDATE_COLOR        = alohssh.UPDATE_COLOR

	SEND_FRIEND_REQ
)

type (
	Event                = alohssh.Event
	FriendConnsData      = alohssh.FriendConnsData
	UsersHardDenoiseData = alohssh.UsersHardDenoiseData
	UsersSoftDenoiseData = alohssh.UsersSoftDenoiseData
	TaglineData          = alohssh.TaglineData
	Identity             = alohssh.Identity
	NicknameData         = alohssh.NicknameData
	ColorData            = alohssh.ColorData
)

func (sc *sshClient) proccessEventsChan() {
	op := "sshClient.proccessEvents"
	log := sc.log.AddOp(op)
	scanner := bufio.NewScanner(sc.eventSSHChannel)
	for scanner.Scan() {
		rawBytes := scanner.Bytes()

		if len(rawBytes) == 0 {
			continue
		}

		e, err := proccessEvent(rawBytes)
		if err != nil {
			log.Error("failed to proccess event", logger.Err(err))
			continue
		}
		eventLog := logger.Attr("event", e)
		select {
		case sc.eventsChan <- e:
			log.Info("new event in events chan", eventLog)
		default:
			log.Error("events chan is full, event skipped", eventLog)
		}

	}

	if err := scanner.Err(); err != nil {
		log.Error("failed to read data stream", logger.Err(err))
	} else {
		log.Info("event processing canceled successfully")
	}
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

func CastToIdentityData(data []byte) (Identity, error) {
	var iden Identity
	if err := json.Unmarshal(data, &iden); err != nil {
		return iden, err
	}

	return iden, nil
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

func CastToColorData(data []byte) (ColorData, error) {
	var cd ColorData
	if err := json.Unmarshal(data, &cd); err != nil {
		return cd, err
	}
	return cd, nil
}

func CastToNickanameData(data []byte) (NicknameData, error) {
	var nd NicknameData
	if err := json.Unmarshal(data, &nd); err != nil {
		return nd, err
	}
	return nd, nil
}

func MarshData(d any) ([]byte, error) {
	data, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	return data, nil
}

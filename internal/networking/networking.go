package networking

import (
	"github.com/google/uuid"
	alohnetwork "github.com/kiryuhakipyatok/aloh-networking"
)

type Networking interface {
	FetchCurrentConnects(id uuid.UUID) ([]uuid.UUID, error)
	FetchCurrentOnline() ([]uuid.UUID, error)
	FetchOnlineFriends(friends []uuid.UUID) (map[uuid.UUID][]string, error)

	ChatCallback(cb func(id uuid.UUID, data []byte))
	VideoCallback(cb func(id uuid.UUID, data []byte))
	VoiceCallback(cb func(id uuid.UUID, data []byte))
	PeerConnectedCallback(cb func(id uuid.UUID))
	PeerDisconnectedCallback(cb func(id uuid.UUID))
	EventCallback(cb func(id uuid.UUID, e alohnetwork.Event))

	SendMessageInChat(msg []byte) error
	SendVoiceData(data []byte) error
	SendVideoData(data []byte) error

	NewEvent(e alohnetwork.Event) error

	ConnectToAllUsers(id uuid.UUID) error
	ConnectToUser(id uuid.UUID) error
	DisconnectFromAllUsers() error
	DisconnectFromUser(id uuid.UUID) error

	Close()
}

type networking struct {
	*alohnetwork.Netwoking
}

func NewNetworking(id uuid.UUID, logPath string) (Networking, error) {
	if len(embeddedConfig) == 0 {
		panic("embedded config is empty")
	}

	cfg := setupConfig()
	cfg.App.LogPath = logPath

	netw, err := alohnetwork.NewNetworking(id, cfg)
	if err != nil {
		return nil, err
	}
	return &networking{netw}, nil
}

func (n *networking) Close() {
	n.Delete()
}

func (n *networking) FetchCurrentConnects(id uuid.UUID) ([]uuid.UUID, error) {
	connects, err := n.FetchSessions(id)
	if err != nil {
		return nil, err
	}
	return connects, nil
}

func (n *networking) FetchCurrentOnline() ([]uuid.UUID, error) {
	online, err := n.FetchOnline()
	if err != nil {
		return nil, err
	}
	return online, nil
}

func (n *networking) FetchOnlineFriends(friends []uuid.UUID) (map[uuid.UUID][]string, error) {
	frs, err := n.FetchFriends(friends)
	if err != nil {
		return nil, err
	}
	return frs, nil
}

func (n *networking) ChatCallback(cb func(id uuid.UUID, data []byte)) {
	n.RegisterOnChat(cb)
}

func (n *networking) VideoCallback(cb func(id uuid.UUID, data []byte)) {
	n.RegisterOnVideo(cb)
}

func (n *networking) VoiceCallback(cb func(id uuid.UUID, data []byte)) {
	n.RegisterOnVoice(cb)
}

func (n *networking) PeerConnectedCallback(cb func(id uuid.UUID)) {
	n.RegisterOnPeerConnected(cb)
}

func (n *networking) PeerDisconnectedCallback(cb func(id uuid.UUID)) {
	n.RegisterOnPeerDisconnected(cb)
}

func (n *networking) EventCallback(cb func(id uuid.UUID, e alohnetwork.Event)) {
	n.RegisterOnEvent(cb)
}

func (n *networking) NewEvent(e alohnetwork.Event) error {
	if err := n.SendEvent(e); err != nil {
		return err
	}
	return nil
}

func (n *networking) SendMessageInChat(msg []byte) error {
	if err := n.SendMessage(msg); err != nil {
		return err
	}
	return nil
}

func (n *networking) SendVoiceData(data []byte) error {
	if err := n.SendVoice(data); err != nil {
		return err
	}
	return nil
}

func (n *networking) SendVideoData(data []byte) error {
	if err := n.SendVideo(data); err != nil {
		return err
	}
	return nil
}

func (n *networking) ConnectToAllUsers(id uuid.UUID) error {
	if err := n.Connect(id); err != nil {
		return err
	}
	return nil
}

func (n *networking) ConnectToUser(id uuid.UUID) error {

	if err := n.ConnectById(id); err != nil {
		return err
	}
	return nil
}

func (n *networking) DisconnectFromAllUsers() error {
	if err := n.Disconnect(); err != nil {
		return err
	}
	return nil
}

func (n *networking) DisconnectFromUser(id uuid.UUID) error {
	if err := n.DisconnectById(id); err != nil {
		return err
	}
	return nil
}

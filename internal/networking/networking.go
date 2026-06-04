package networking

import (
	alohnetwork "github.com/kiryuhakipyatok/aloh-networking"
)

type Networking interface {
	FetchCurrentConnects(nickname string) ([]string, error)
	FetchCurrentOnline() ([]string, error)
	FetchOnlineFriends(friends []string) (map[string][]string, error)

	ChatCallback(cb func(id string, data []byte))
	VideoCallback(cb func(id string, data []byte))
	VoiceCallback(cb func(id string, data []byte))
	PeerConnectedCallback(cb func(id string))
	PeerDisconnectedCallback(cb func(id string))

	SendMessageInChat(msg []byte) error
	SendVoiceData(data []byte) error
	SendVideoData(data []byte) error

	ConnectToAllUsers(nickname string) error
	ConnectToUser(nickname string) error
	DisconnectFromAllUsers() error
	DisconnectFromUser(id string) error

	Close()
}

type networking struct {
	*alohnetwork.Netwoking
}

func NewNetworking(nickname, logPath string) (Networking, error) {
	if len(embeddedConfig) == 0 {
		panic("embedded config is empty")
	}

	cfg := setupConfig()
	cfg.App.LogPath = logPath
	netw, err := alohnetwork.NewNetworking(nickname, cfg)
	if err != nil {
		return nil, err
	}
	return &networking{netw}, nil
}

func (n *networking) Close() {
	n.Delete()
}

func (n *networking) FetchCurrentConnects(nickname string) ([]string, error) {
	connects, err := n.FetchSessions(nickname)
	if err != nil {
		return nil, err
	}
	return connects, nil
}

func (n *networking) FetchCurrentOnline() ([]string, error) {
	online, err := n.FetchOnline()
	if err != nil {
		return nil, err
	}
	return online, nil
}

func (n *networking) FetchOnlineFriends(friends []string) (map[string][]string, error) {
	frs, err := n.FetchFriends(friends)
	if err != nil {
		return nil, err
	}
	return frs, nil
}

func (n *networking) ChatCallback(cb func(id string, data []byte)) {
	n.RegisterOnChat(cb)
}

func (n *networking) VideoCallback(cb func(id string, data []byte)) {
	n.RegisterOnVideo(cb)
}

func (n *networking) VoiceCallback(cb func(id string, data []byte)) {
	n.RegisterOnVoice(cb)
}

func (n *networking) PeerConnectedCallback(cb func(id string)) {
	n.RegisterOnPeerConnected(cb)
}

func (n *networking) PeerDisconnectedCallback(cb func(id string)) {
	n.RegisterOnPeerDisconnected(cb)
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

func (n *networking) ConnectToAllUsers(nickname string) error {
	if err := n.Connect(nickname); err != nil {
		return err
	}
	return nil
}

func (n *networking) ConnectToUser(nickname string) error {
	if err := n.ConnectById(nickname); err != nil {
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

func (n *networking) DisconnectFromUser(id string) error {
	if err := n.DisconnectById(id); err != nil {
		return err
	}
	return nil
}

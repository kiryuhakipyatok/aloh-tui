package networking

import (
	networkapi "github.com/kiryuhakipyatok/aloh-networking/cmd/api"
)

type Networking interface {
	FetchCurrentConnects(nickname string) ([]string, error)
	FetchCurrentOnline() ([]string, error)

	ChatCallback(cb func(id string, data []byte))
	VideoCallback(cb func(id string, data []byte))
	VoiceCallback(cb func(id string, data []byte))

	SendMessageInChat(msg string) error
	SendVoiceData(data []byte) error
	SendVideoData(data []byte) error

	ConnectToUser(nickname string) error
	DisconnectFromUsers() error

	Close()
}

type networking struct {
	*networkapi.Netwoking
}

func NewNetworking(nickname, logPath string) (Networking, error) {
	netw, err := networkapi.NewNetworking(nickname, logPath)
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

func (n *networking) ChatCallback(cb func(id string, data []byte)) {
	n.RegisterOnChat(cb)
}

func (n *networking) VideoCallback(cb func(id string, data []byte)) {
	n.RegisterOnVideo(cb)
}

func (n *networking) VoiceCallback(cb func(id string, data []byte)) {
	n.RegisterOnVoice(cb)
}

func (n *networking) SendMessageInChat(msg string) error {
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

func (n *networking) ConnectToUser(nickname string) error {
	if err := n.Connect(nickname); err != nil {
		return err
	}
	return nil
}

func (n *networking) DisconnectFromUsers() error {
	if err := n.Disconnect(); err != nil {
		return err
	}
	return nil
}

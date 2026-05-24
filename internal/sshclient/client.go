package sshclient

import (
	"aloh-tui/pkg/errs"
	"context"
	"errors"
	"fmt"
	"io"

	"aloh-tui/pkg/logger"

	"golang.org/x/crypto/ssh"
	cssh "golang.org/x/crypto/ssh"
)

const (
	host = "164.90.163.153"
	port = "48713"

	REGISTER = iota
	LOGIN
	DEFAULT
)

const (
	SUCCESS = iota
	NOT_FOUND
	ALREADY_EXISTS
	SERVER_ERROR
)

type SSHClient interface {
	NewFriendReq(ctx context.Context, friendNickname string) error
	AcceptFriendReq(ctx context.Context, friendNickname string) error
	DenyFriendReq(ctx context.Context, friendNickname string) error
	DeleteFromFriends(ctx context.Context, friendNickname string) error
	Close()
}

type sshClient struct {
	client          *cssh.Client
	eventSSHChannel cssh.Channel
	eventsChan      chan Event
	log             *logger.Logger
}

type SSHClientSetup struct {
	Nickname   string
	KeysPath   string
	Typee      uint
	Password   []byte
	EventsChan chan Event
}

func AuthSSHClient(ctx context.Context, l *logger.Logger, setup SSHClientSetup) (SSHClient, []byte, error) {
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
		var (
			payload  []byte
			status   bool
			authType string
			errAuth  error
		)

		switch setup.Typee {
		case LOGIN:
			authType = "SSH-2.0-aloh-login"
			errAuth = errs.ErrLogin
		case REGISTER:
			authType = "SSH-2.0-aloh-register"
			errAuth = errs.ErrRegister
		case DEFAULT:
			authType = "SSH-2.0-aloh-default"
			errAuth = errs.ErrAuth
		default:
			authType = "SSH-2.0-aloh-default"
			errAuth = errs.ErrAuth
		}

		kp, err := InitKeys(setup.KeysPath)
		if err != nil {
			return nil, nil, err
		}

		addr := fmt.Sprintf("%s:%s", host, port)
		signer, err := cssh.ParsePrivateKey(kp.RawPrivateKey())
		if err != nil {
			return nil, nil, err
		}
		client, err := cssh.Dial("tcp", addr, &cssh.ClientConfig{
			User: setup.Nickname,
			Auth: []cssh.AuthMethod{
				cssh.PublicKeys(signer),
				cssh.Password(string(setup.Password)),
			},
			HostKeyCallback: cssh.InsecureIgnoreHostKey(),
			ClientVersion:   authType,
		})
		if err != nil {
			if authErr(err.Error()) {
				return nil, nil, errAuth
			}
			return nil, nil, err
		}

		keyBytes := kp.RawAuthorizedKey()

		switch setup.Typee {
		case LOGIN:
			status, _, err = client.SendRequest("key", true, keyBytes)
			if err != nil {
				return nil, nil, err
			}
			if !status {
				return nil, nil, castErr(payload)
			}
		case REGISTER:
			status, _, err = client.SendRequest("pswrd", true, setup.Password)
			if err != nil {
				return nil, nil, err
			}
			if !status {
				return nil, nil, castErr(payload)
			}
		default:
		}
		if setup.Typee != REGISTER {
			status, payload, err = client.SendRequest("personal-data", true, nil)
			if err != nil {
				return nil, nil, err
			}
			if !status {
				return nil, nil, castErr(payload)
			}
		}
		eventChannel, requests, err := client.OpenChannel("event-channel", nil)
		if err != nil {
			return nil, nil, castErr(payload)
		}
		go ssh.DiscardRequests(requests)

		sshClient := sshClient{
			client:          client,
			eventSSHChannel: eventChannel,
			eventsChan:      setup.EventsChan,
			log:             l,
		}
		go sshClient.proccessEventsChan()
		return &sshClient, payload, nil
	}

}

func (sc *sshClient) proccessEventsChan() {
	op := "sshClient.proccessEvents"
	log := sc.log.AddOp(op)
	data := make([]byte, 1024)
	for {
		n, err := sc.eventSSHChannel.Read(data)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Info("events processing canceled")
				break
			}
			log.Error("failed to read data", logger.Err(err))
			continue
		}
		if len(data[:n]) > 0 {
			e, err := proccessEvent(data[:n])
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
	}
}

func (sc *sshClient) NewFriendReq(ctx context.Context, friendNickname string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		status, payload, err := sc.client.SendRequest("new-friend", true, []byte(friendNickname))
		if err != nil {
			return err
		}
		if !status {
			return castErr(payload)
		}
		return nil
	}
}

func (sc *sshClient) AcceptFriendReq(ctx context.Context, friendNickname string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		status, payload, err := sc.client.SendRequest("accept-friend", true, []byte(friendNickname))
		if err != nil {
			return err
		}
		if !status {
			return castErr(payload)
		}
		return nil
	}
}

func (sc *sshClient) DenyFriendReq(ctx context.Context, friendNickname string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		status, payload, err := sc.client.SendRequest("deny-friend", true, []byte(friendNickname))
		if err != nil {
			return err
		}
		if !status {
			return castErr(payload)
		}
		return nil
	}
}

func (sc *sshClient) DeleteFromFriends(ctx context.Context, friendNickname string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		status, payload, err := sc.client.SendRequest("delete-friend", true, []byte(friendNickname))
		if err != nil {
			return err
		}
		if !status {
			return castErr(payload)
		}
		return nil
	}
}

func (sc *sshClient) Close() {
	sc.log.Info("closing ssh client")
	if err := sc.eventSSHChannel.Close(); err != nil {
		sc.log.Error("failed to close event channel", logger.Err(err))
	}
	if err := sc.client.Close(); err != nil {
		sc.log.Error("failed to close ssh client", logger.Err(err))
	}
}

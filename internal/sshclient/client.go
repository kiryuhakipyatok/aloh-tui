package sshclient

import (
	"aloh-tui/pkg/errs"
	"context"
	"fmt"

	"aloh-tui/pkg/logger"

	"github.com/google/uuid"
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

type SSHClient interface {
	NewFriendReq(ctx context.Context, nickname string) error
	AcceptFriendReq(ctx context.Context, iden Identity) error
	DenyFriendReq(ctx context.Context, id uuid.UUID) error
	DeleteFromFriends(ctx context.Context, id uuid.UUID) error
	BlockUser(ctx context.Context, nickname string) ([]byte, error)
	UnblockUser(ctx context.Context, nickname string) ([]byte, error)
	UpdateCurrentConnects(ctx context.Context, conns []Identity) error
	SetTagline(ctx context.Context, tagline string) error
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
			errAuth = errs.ErrLogin()
		case REGISTER:
			authType = "SSH-2.0-aloh-register"
			errAuth = errs.ErrRegister()
		case DEFAULT:
			authType = "SSH-2.0-aloh-default"
			errAuth = errs.ErrAuth()
		default:
			authType = "SSH-2.0-aloh-default"
			errAuth = errs.ErrAuth()
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
		//if setup.Typee != REGISTER {
		status, payload, err = client.SendRequest("personal-data", true, nil)
		if err != nil {
			return nil, nil, err
		}
		if !status {
			return nil, nil, castErr(payload)
		}
		l.Info("pd", string(payload))
		//}
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

func (sc *sshClient) Close() {
	sc.log.Info("closing ssh client")
	if err := sc.eventSSHChannel.Close(); err != nil {
		sc.log.Error("failed to close event channel", logger.Err(err))
	}
	if err := sc.client.Close(); err != nil {
		sc.log.Error("failed to close ssh client", logger.Err(err))
	}
}

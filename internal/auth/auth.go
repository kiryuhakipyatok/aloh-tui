package auth

import (
	"aloh-tui/internal/auth/ssh"
	"aloh-tui/pkg/errs"
	"fmt"
	"strings"

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

func Auth(nickname, keysPath string, typee uint, password []byte) ([]byte, error) {
	var (
		payload  []byte
		status   bool
		authType string
		errAuth  error
	)

	switch typee {
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

	kp, err := ssh.InitKeys(keysPath)
	if err != nil {
		return nil, err
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	signer, err := cssh.ParsePrivateKey(kp.RawPrivateKey())
	if err != nil {
		return nil, err
	}
	client, err := cssh.Dial("tcp", addr, &cssh.ClientConfig{
		User: nickname,
		Auth: []cssh.AuthMethod{
			cssh.PublicKeys(signer),
			cssh.Password(string(password)),
		},
		HostKeyCallback: cssh.InsecureIgnoreHostKey(),
		ClientVersion:   authType,
	})
	if err != nil {
		if authErr(err.Error()) {
			return nil, errAuth
		}
		return nil, err
	}

	keyBytes := kp.RawAuthorizedKey()

	switch typee {
	case LOGIN:
		status, payload, err = client.SendRequest("key", true, keyBytes)
		if err != nil {
			return nil, errAuth
		}
		if !status {
			return nil, castErr(payload)
		}
	case REGISTER:
		status, payload, err = client.SendRequest("pswrd", true, password)
		if err != nil {
			return nil, errAuth
		}
		if !status {
			return nil, castErr(payload)
		}
	default:
	}

	if err := client.Close(); err != nil {
		return nil, err
	}

	return payload, nil
}

func authErr(errStr string) bool {
	return strings.Contains(errStr, "unable to authenticate")
}

func castErr(errByte []byte) error {
	switch errByte[0] {
	case NOT_FOUND:
		return errs.ErrNotFound
	case ALREADY_EXISTS:
		return errs.ErrAlreadyExists
	default:
		return errs.ErrInternalServer
	}
}

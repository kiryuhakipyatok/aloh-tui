package sshclient

import (
	"aloh-tui/pkg/errs"
	"strings"
)

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

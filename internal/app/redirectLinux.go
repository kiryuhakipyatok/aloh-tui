//go:build !windows

package app

import (
	"io"
	l "log"
	"os"
	"syscall"
)

func redirectErr(crashLogPath string) (*os.File, error) {
	null, err := os.OpenFile(crashLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	os.Stderr = null
	syscall.Dup2(int(null.Fd()), 2)
	l.SetOutput(io.Discard)
	return null, err
}

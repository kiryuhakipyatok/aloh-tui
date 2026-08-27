//go:build !windows

package app

import (
	"io"
	l "log"
	"os"
	"syscall"
)

func redirectErr(errorsLogPath string) (*os.File, error) {
	errorsFile, err := os.OpenFile(errorsLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	os.Stderr = errorsFile
	syscall.Dup2(int(errorsFile.Fd()), 2)
	l.SetOutput(io.Discard)
	return errorsFile, err
}

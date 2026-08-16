//go:build windows

package app

import (
	"io"
	l "log"
	"os"
)

func redirectErr(crashLogPath string) (*os.File, error) {
	null, err := os.OpenFile(crashLogPath, os.O_CREATE|os.O_WRONLY, 0)
	if err != nil {
		return nil, err
	}
	os.Stderr = null
	l.SetOutput(io.Discard)
	return null, err
}
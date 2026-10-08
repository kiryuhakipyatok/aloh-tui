//go:build !windows

package app

import (
	"io"
	l "log"
	"os"
	"os/exec"
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

func udpSizeSetup() error {
	cmd1 := exec.Command("sysctl", "-w", "net.core.rmem_max=7500000")
	cmd2 := exec.Command("sysctl", "-w", "net.core.wmem_max=7500000")
	if err := cmd1.Run(); err != nil {
		return err
	}
	if err := cmd2.Run(); err != nil {
		return err
	}
	return nil
}

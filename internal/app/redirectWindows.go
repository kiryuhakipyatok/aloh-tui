//go:build windows

package app

import (
	"os"
	"syscall"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	setStdHandleProc = kernel32.NewProc("SetStdHandle")
)

const stdErrorHandle = syscall.STD_ERROR_HANDLE

func redirectErr(errorsLogPath string) (*os.File, error) {
	errorsFile, err := os.OpenFile(errorsLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	os.Stderr = errorsFile

	err = setStdHandle(stdErrorHandle, syscall.Handle(errorsFile.Fd()))
	if err != nil {
		errorsFile.Close()
		return nil, err
	}

	return errorsFile, nil
}

func setStdHandle(stdHandle int, handle syscall.Handle) error {
	r0, _, e1 := syscall.SyscallN(setStdHandleProc.Addr(), uintptr(stdHandle), uintptr(handle))
	if r0 == 0 {
		if e1 != 0 {
			return error(e1)
		}
		return syscall.EINVAL
	}
	return nil
}

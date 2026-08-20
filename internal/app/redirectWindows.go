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

func redirectErr(crashLogPath string) (*os.File, error) {
	file, err := os.OpenFile(crashLogPath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}

	os.Stderr = file

	err = setStdHandle(stdErrorHandle, syscall.Handle(file.Fd()))
	if err != nil {
		file.Close()
		return nil, err
	}

	return file, nil
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

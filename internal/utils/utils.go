package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type FilesPaths struct {
	NetworkingLog string
	AppLog        string
	KeysDir       string
	UserdataFile  string
	CrashLog      string
}

func SetupFiles() (FilesPaths, error) {
	binPath, err := os.Executable()
	if err != nil {
		return FilesPaths{}, err
	}

	binDir := filepath.Dir(binPath)

	netwLogPath := filepath.Join(binDir, "networking-logs")
	appLogPath := filepath.Join(binDir, "app-logs")
	keysPath := filepath.Join(binDir, "keys")
	userdataFilePath := filepath.Join(binDir, "userdata.json")
	crashLogFilePath := filepath.Join(binDir, "crash.log")

	if err := os.MkdirAll(netwLogPath, 0755); err != nil {
		return FilesPaths{}, err
	}

	if err := os.MkdirAll(appLogPath, 0755); err != nil {
		return FilesPaths{}, err
	}

	netwLogFilePath := filepath.Join(netwLogPath, "log.log")
	appLogFilePath := filepath.Join(appLogPath, "log.log")

	newtLogFile, err := os.OpenFile(netwLogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, 0644)
	if err != nil {
		return FilesPaths{}, err
	}
	if err := newtLogFile.Close(); err != nil {
		return FilesPaths{}, err
	}

	appLogFile, err := os.OpenFile(appLogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, 0644)
	if err != nil {
		return FilesPaths{}, err
	}
	if err := appLogFile.Close(); err != nil {
		return FilesPaths{}, err
	}

	if err := os.MkdirAll(keysPath, 0700); err != nil {
		return FilesPaths{}, err
	}

	nickFile, err := os.OpenFile(userdataFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return FilesPaths{}, err
	}

	if err := nickFile.Close(); err != nil {
		return FilesPaths{}, err
	}

	return FilesPaths{
		NetworkingLog: netwLogFilePath,
		AppLog:        appLogFilePath,
		UserdataFile:  userdataFilePath,
		KeysDir:       keysPath,
		CrashLog:      crashLogFilePath,
	}, nil
}

func DarkenHex(hex string, factor float64) string {
	hex = strings.TrimPrefix(hex, "#")

	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}

	if len(hex) != 6 {
		return ""
	}

	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return ""
	}
	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return ""
	}
	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return ""
	}

	darken := func(color uint64) uint64 {
		val := float64(color) * factor
		if val < 0 {
			return 0
		}
		if val > 255 {
			return 255
		}
		return uint64(val)
	}

	return fmt.Sprintf("#%02X%02X%02X", darken(r), darken(g), darken(b))
}

func SetThreeFirstByte(threeFb []byte, data []byte) []byte {
	new := make([]byte, 0, len(data)+3)
	new = append(new, threeFb...)
	new = append(new, data...)
	return new
}

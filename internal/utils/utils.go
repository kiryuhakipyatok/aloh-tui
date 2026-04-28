package utils

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/x/ansi"
)

func Difference(a, b []string) []string {
	bMap := make(map[string]struct{})
	for _, item := range b {
		bMap[ansi.Strip(item)] = struct{}{}
	}

	var diff []string

	for _, item := range a {

		if _, ok := bMap[ansi.Strip(item)]; !ok {
			diff = append(diff, item)
		}
	}

	return diff
}

func SetupFiles() (nlP, alP, nP, kP string, err error) {
	binPath, err := os.Executable()
	if err != nil {
		return "", "", "", "", err
	}

	binDir := filepath.Dir(binPath)

	netwLogPath := filepath.Join(binDir, "networking-logs")
	appLogPath := filepath.Join(binDir, "app-logs")
	keysPath := filepath.Join(binDir, "keys")
	nickFilePath := filepath.Join(binDir, "userdata.json")

	if err := os.MkdirAll(netwLogPath, 0755); err != nil {
		return "", "", "", "", err
	}

	if err := os.MkdirAll(appLogPath, 0755); err != nil {
		return "", "", "", "", err
	}

	netwLogFilePath := filepath.Join(netwLogPath, "log.log")
	appLogFilePath := filepath.Join(appLogPath, "log.log")

	newtLogFile, err := os.OpenFile(netwLogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, 0644)
	if err != nil {
		return "", "", "", "", err
	}
	if err := newtLogFile.Close(); err != nil {
		return "", "", "", "", err
	}

	appLogFile, err := os.OpenFile(appLogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, 0644)
	if err != nil {
		return "", "", "", "", err
	}
	if err := appLogFile.Close(); err != nil {
		return "", "", "", "", err
	}

	if err := os.MkdirAll(keysPath, 0700); err != nil {
		return "", "", "", "", err
	}

	nickFile, err := os.OpenFile(nickFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return "", "", "", "", err
	}

	if err := nickFile.Close(); err != nil {
		return "", "", "", "", err
	}

	return netwLogFilePath, appLogFilePath, nickFilePath, keysPath, nil
}


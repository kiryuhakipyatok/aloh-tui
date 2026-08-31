package sshclient

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/keygen"
)

func InitKeys(keysDir string, typee uint) (*keygen.KeyPair, error) {
	if typee != DEFAULT {
		if err := os.RemoveAll(keysDir); err != nil {
			return nil, err
		}
	}

	kp, err := keygen.New(filepath.Join(keysDir, "id_ed25519"), keygen.WithKeyType(keygen.Ed25519), keygen.WithWrite())
	if err != nil {
		return nil, err
	}
	return kp, nil
}

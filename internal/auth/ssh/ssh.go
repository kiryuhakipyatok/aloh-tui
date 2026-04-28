package ssh

import (
	"path/filepath"

	"github.com/charmbracelet/keygen"
)

func InitKeys(keysDir string) (*keygen.KeyPair, error) {
	kp, err := keygen.New(filepath.Join(keysDir, "id_ed25519"), keygen.WithKeyType(keygen.Ed25519), keygen.WithWrite())
	if err != nil {
		return nil, err
	}
	return kp, nil
}
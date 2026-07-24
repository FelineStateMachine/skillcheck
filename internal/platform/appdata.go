package platform

import (
	"os"
	"path/filepath"
)

func AppDataDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "skilltrace"), nil
}

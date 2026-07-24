package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Store struct{ Path string }

func (s Store) Load() (Settings, error) {
	b, err := os.ReadFile(s.Path)
	if os.IsNotExist(err) {
		return Settings{Version: 1}, nil
	}
	if err != nil {
		return Settings{}, err
	}
	var settings Settings
	if err := json.Unmarshal(b, &settings); err != nil {
		return Settings{}, fmt.Errorf("decode settings: %w", err)
	}
	if settings.Version != 1 {
		return Settings{}, fmt.Errorf("unsupported settings version")
	}
	for i, root := range settings.Roots {
		validated, err := ValidateRoot(root)
		if err != nil {
			return Settings{}, fmt.Errorf("invalid root %d: %w", i, err)
		}
		settings.Roots[i] = validated
	}
	return settings, nil
}

func (s Store) Save(settings Settings) error {
	settings.Version = 1
	for _, root := range settings.Roots {
		if _, err := ValidateRoot(root); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.Path), ".settings-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err = tmp.Chmod(0o600); err == nil {
		_, err = tmp.Write(append(data, '\n'))
	}
	if err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(name, s.Path)
}

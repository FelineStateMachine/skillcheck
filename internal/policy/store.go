package policy

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

type Store struct{}

func (Store) Read(path string) (*Document, []Diagnostic, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	doc, diagnostics := Parse(data)
	return doc, diagnostics, nil
}

func (Store) Replace(path, expectedFingerprint string, data []byte) (string, error) {
	current, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(current)
	actual := fmt.Sprintf("%x", sum[:])
	if actual != expectedFingerprint {
		return "", ConflictError{Message: "policy changed since preview"}
	}
	if _, diagnostics := Parse(data); len(diagnostics) > 0 {
		return "", diagnostics[0]
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".skilltrace-policy-*")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(info.Mode().Perm()); err != nil {
		tmp.Close()
		return "", err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return "", err
	}
	updated := sha256.Sum256(data)
	return fmt.Sprintf("%x", updated[:]), nil
}

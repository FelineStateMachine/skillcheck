package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
)

var ErrStale = errors.New("evidence source changed")
var ErrUnavailable = errors.New("evidence source unavailable")

type Reference struct {
	Token, SourceFingerprint string
	Offset, Length           int64
}
type Result struct {
	Token, State string
	Content      []byte
}
type Resolver struct{}

func Fingerprint(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func (Resolver) Resolve(path string, ref Reference) (Result, error) {
	fingerprint, err := Fingerprint(path)
	if err != nil {
		return Result{Token: ref.Token, State: "unavailable"}, ErrUnavailable
	}
	if fingerprint != ref.SourceFingerprint {
		return Result{Token: ref.Token, State: "stale"}, ErrStale
	}
	f, err := os.Open(path)
	if err != nil {
		return Result{Token: ref.Token, State: "unavailable"}, ErrUnavailable
	}
	defer f.Close()
	if _, err = f.Seek(ref.Offset, io.SeekStart); err != nil {
		return Result{Token: ref.Token, State: "unavailable"}, ErrUnavailable
	}
	b := make([]byte, ref.Length)
	n, err := io.ReadFull(f, b)
	if err != nil && err != io.ErrUnexpectedEOF {
		return Result{Token: ref.Token, State: "unavailable"}, ErrUnavailable
	}
	return Result{Token: ref.Token, State: "available", Content: b[:n]}, nil
}

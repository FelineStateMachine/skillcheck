package source

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

type Fingerprint struct {
	Size         int64  `json:"size"`
	ModTimeNanos int64  `json:"mod_time_nanos"`
	Identity     string `json:"identity"`
	PrefixDigest string `json:"prefix_digest"`
	WindowDigest string `json:"window_digest"`
}

func FingerprintFile(path string, committedOffset int64) (Fingerprint, error) {
	f, err := os.Open(path)
	if err != nil {
		return Fingerprint{}, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return Fingerprint{}, err
	}
	prefixEnd := committedOffset
	if prefixEnd > info.Size() {
		prefixEnd = info.Size()
	}
	if prefixEnd < 0 {
		prefixEnd = 0
	}
	prefix := sha256.New()
	if _, err := io.CopyN(prefix, f, prefixEnd); err != nil && err != io.EOF {
		return Fingerprint{}, err
	}
	windowStart := prefixEnd - 4096
	if windowStart < 0 {
		windowStart = 0
	}
	if _, err := f.Seek(windowStart, io.SeekStart); err != nil {
		return Fingerprint{}, err
	}
	window := sha256.New()
	if _, err := io.CopyN(window, f, prefixEnd-windowStart); err != nil && err != io.EOF {
		return Fingerprint{}, err
	}
	return Fingerprint{Size: info.Size(), ModTimeNanos: info.ModTime().UnixNano(), Identity: fileIdentity(info), PrefixDigest: hex.EncodeToString(prefix.Sum(nil)), WindowDigest: hex.EncodeToString(window.Sum(nil))}, nil
}

func fileIdentity(info os.FileInfo) string {
	// FileInfo.Sys differs by platform; a stable metadata digest avoids exposing paths.
	h := sha256.Sum256([]byte(info.Name() + "\x00" + info.Mode().String()))
	return hex.EncodeToString(h[:12])
}

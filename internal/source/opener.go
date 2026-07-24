package source

import (
	"bufio"
	"bytes"
	"io"
	"os"
)

// OpenSuffix returns complete newline-terminated JSONL records after a checkpoint.
// An incomplete trailing record is deliberately left for the next refresh.
func OpenSuffix(path string, offset int64) (io.ReadCloser, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, offset, err
	}
	if _, err = f.Seek(offset, io.SeekStart); err != nil {
		f.Close()
		return nil, offset, err
	}
	r := bufio.NewReader(f)
	var complete bytes.Buffer
	for {
		line, readErr := r.ReadBytes('\n')
		if len(line) > 0 && line[len(line)-1] == '\n' {
			complete.Write(line)
		}
		if readErr != nil {
			break
		}
	}
	if err := f.Close(); err != nil {
		return nil, offset, err
	}
	return io.NopCloser(bytes.NewReader(complete.Bytes())), offset + int64(complete.Len()), nil
}

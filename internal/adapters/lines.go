package adapters

import (
	"bufio"
	"errors"
	"io"
)

// ErrIncompleteTrailingRecord signals that a trace ended in the middle of a
// record: a non-empty final line with no terminating newline that the adapter
// could not parse. That is the fingerprint of a file captured mid-write, and
// committing it would persist half a session, so adapters surface it as a hard
// error rather than an exclusion. A malformed line anywhere else is recoverable
// and should be excluded instead.
var ErrIncompleteTrailingRecord = errors.New("incomplete trailing record")

// ForEachLine calls fn for every newline-delimited record in r. terminated is
// false only for a final line that carried no newline, letting the adapter tell
// a mid-write truncation from an ordinary bad line.
func ForEachLine(r io.Reader, maxBytes int, fn func(lineNo int64, data []byte, terminated bool) error) error {
	br := bufio.NewReaderSize(r, 64*1024)
	var lineNo int64
	for {
		data, err := readLine(br, maxBytes)
		if len(data) > 0 {
			lineNo++
			terminated := data[len(data)-1] == '\n'
			trimmed := data
			if terminated {
				trimmed = data[:len(data)-1]
				if n := len(trimmed); n > 0 && trimmed[n-1] == '\r' {
					trimmed = trimmed[:n-1]
				}
			}
			if len(trimmed) > 0 {
				if cbErr := fn(lineNo, trimmed, terminated); cbErr != nil {
					return cbErr
				}
			}
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

// readLine returns one line including its trailing newline, or the final
// unterminated remainder. It bounds the line length so a pathological input
// cannot exhaust memory.
func readLine(br *bufio.Reader, maxBytes int) ([]byte, error) {
	var buf []byte
	for {
		chunk, err := br.ReadSlice('\n')
		buf = append(buf, chunk...)
		if len(buf) > maxBytes {
			return buf[:maxBytes], io.EOF
		}
		if err == bufio.ErrBufferFull {
			continue
		}
		return buf, err
	}
}

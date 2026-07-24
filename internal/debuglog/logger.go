package debuglog

import (
	"fmt"
	"io"
	"sync"
)

// Logger is an opt-in, bounded diagnostic sink. Callers must pass sanitized messages.
type Logger struct {
	mu        sync.Mutex
	w         io.Writer
	remaining int64
}

func New(w io.Writer, limit int64) *Logger { return &Logger{w: w, remaining: limit} }

func (l *Logger) Printf(format string, args ...any) {
	if l == nil || l.w == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.remaining <= 0 {
		return
	}
	b := []byte(fmt.Sprintf(format, args...) + "\n")
	if int64(len(b)) > l.remaining {
		b = b[:l.remaining]
	}
	n, _ := l.w.Write(b)
	l.remaining -= int64(n)
}

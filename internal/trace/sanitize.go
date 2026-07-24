package trace

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

type Sanitizer interface {
	Token(category, raw string) string
	Label(raw string) string
}

type HashSanitizer struct{ salt string }

func NewHashSanitizer(salt string) HashSanitizer { return HashSanitizer{salt: salt} }

func (s HashSanitizer) Token(category, raw string) string {
	h := sha256.Sum256([]byte(s.salt + "\x00" + category + "\x00" + raw))
	return category + ":" + hex.EncodeToString(h[:8])
}

func (s HashSanitizer) Label(raw string) string {
	var b strings.Builder
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		}
	}
	if b.Len() > 64 {
		return b.String()[:64]
	}
	return b.String()
}

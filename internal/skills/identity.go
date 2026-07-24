package skills

import "crypto/sha256"

type Identity struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Revision    string `json:"revision"`
}

func identity(name, description, body string) Identity {
	s := sha256.Sum256([]byte(name + "\x00" + body))
	r := sha256.Sum256([]byte(body))
	return Identity{ID: "skill-" + hex(s[:8]), Name: name, Description: description, Revision: hex(r[:8])}
}

func hex(b []byte) string {
	const h = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, v := range b {
		out[i*2], out[i*2+1] = h[v>>4], h[v&15]
	}
	return string(out)
}

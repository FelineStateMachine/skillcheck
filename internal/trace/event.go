package trace

import (
	"encoding/json"
	"fmt"
)

const ContractVersion = 1

type Coordinate struct {
	Line int64 `json:"line"`
}

type Event struct {
	Version    int             `json:"version"`
	Kind       string          `json:"kind"`
	Sequence   int64           `json:"sequence"`
	Coordinate Coordinate      `json:"coordinate"`
	Payload    json.RawMessage `json:"payload"`
}

func NewEvent(kind string, sequence, line int64, payload any) (Event, error) {
	switch kind {
	case "session", "skill", "tool", "file", "completion", "usage":
	default:
		return Event{}, fmt.Errorf("unsupported event kind %q", kind)
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return Event{}, fmt.Errorf("marshal payload: %w", err)
	}
	return Event{Version: ContractVersion, Kind: kind, Sequence: sequence, Coordinate: Coordinate{Line: line}, Payload: b}, nil
}

func DecodePayload(event Event, target any) error { return json.Unmarshal(event.Payload, target) }

package huggingface

import "encoding/json"

type row struct {
	Harness     string          `json:"harness"`
	RawTrace    json.RawMessage `json:"raw_trace"`
	RawRetained bool            `json:"raw_retained"`
}

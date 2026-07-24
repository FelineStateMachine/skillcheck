package headless

const SchemaVersion = 1

// Progress is the stable stderr NDJSON contract for long-running commands.
type Progress struct {
	Version   int    `json:"version"`
	Type      string `json:"type"`
	Command   string `json:"command"`
	Stage     string `json:"stage"`
	Completed int    `json:"completed"`
	Total     int    `json:"total"`
}

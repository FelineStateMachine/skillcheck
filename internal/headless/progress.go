package headless

import (
	"encoding/json"
	"io"
)

func RenderProgress(w io.Writer, command, stage string, completed, total int) error {
	return json.NewEncoder(w).Encode(Progress{Version: SchemaVersion, Type: "progress", Command: command, Stage: stage, Completed: completed, Total: total})
}

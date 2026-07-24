package headless

import (
	"encoding/json"
	"fmt"
	"io"
)

func Render(w io.Writer, format string, result Envelope) error {
	if format == "json" {
		e := json.NewEncoder(w)
		e.SetEscapeHTML(true)
		return e.Encode(result)
	}
	if format != "text" {
		return fmt.Errorf("unsupported format")
	}
	if !result.OK {
		_, err := fmt.Fprintf(w, "error [%s]: %s\n", result.Error.Code, result.Error.Message)
		return err
	}
	b, err := json.MarshalIndent(result.Data, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s\n", b)
	return err
}

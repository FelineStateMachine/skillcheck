package headless

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
type Envelope struct {
	Version int          `json:"version"`
	OK      bool         `json:"ok"`
	Command string       `json:"command"`
	Data    any          `json:"data,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
}

func Success(command string, data any) Envelope {
	return Envelope{Version: SchemaVersion, OK: true, Command: command, Data: data}
}
func Failure(command, code, message string) Envelope {
	return Envelope{Version: SchemaVersion, OK: false, Command: command, Error: &ErrorDetail{Code: code, Message: message}}
}

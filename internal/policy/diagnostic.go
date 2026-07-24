package policy

import "fmt"

type Diagnostic struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
}

func (d Diagnostic) Error() string {
	if d.Line > 0 {
		return fmt.Sprintf("%s at %d:%d", d.Message, d.Line, d.Column)
	}
	return d.Message
}

package workflow

import "skilltrace/internal/trace"

// Class groups tools by what they do, so a diagram can read as inspect →
// modify → execute even when the concrete tools differ. The concrete tool name
// is still kept on the node; the class is the coarse grouping.
type Class string

const (
	ClassInspect  Class = "inspect"  // read the world: Read, Grep, Glob, WebFetch
	ClassModify   Class = "modify"   // change files: Edit, Write, NotebookEdit
	ClassExecute  Class = "execute"  // run things: Bash, shell, exec_command
	ClassDelegate Class = "delegate" // hand off: Agent, Task
	ClassConverse Class = "converse" // ask the user: AskUserQuestion
	ClassSkill    Class = "skill"    // a skill boundary
	ClassContext  Class = "context"  // session, usage, completion
	ClassOther    Class = "other"
)

var toolClasses = map[string]Class{
	"Read": ClassInspect, "Grep": ClassInspect, "Glob": ClassInspect,
	"WebFetch": ClassInspect, "WebSearch": ClassInspect, "ToolSearch": ClassInspect,
	"Edit": ClassModify, "Write": ClassModify, "NotebookEdit": ClassModify, "MultiEdit": ClassModify,
	"apply_patch": ClassModify,
	"Bash":        ClassExecute, "shell": ClassExecute, "exec_command": ClassExecute,
	"Agent": ClassDelegate, "Task": ClassDelegate, "TaskCreate": ClassDelegate, "TaskUpdate": ClassDelegate,
	"AskUserQuestion": ClassConverse,
}

// Action is the node label for an event: the concrete tool name when there is
// one, falling back to the event kind. This is what turned the old graph — six
// nodes, every tool collapsed into one labelled "tool" — into one node per
// distinct tool.
func Action(event trace.Event) string {
	if event.Kind == "tool" {
		var p trace.ToolPayload
		if trace.DecodePayload(event, &p) == nil && p.Tool != "" {
			return p.Tool
		}
	}
	return event.Kind
}

// Classify assigns the semantic class of an event.
func classify(event trace.Event) Class {
	switch event.Kind {
	case "skill":
		return ClassSkill
	case "session", "usage", "completion":
		return ClassContext
	case "file":
		return ClassModify
	case "tool":
		if c, ok := toolClasses[Action(event)]; ok {
			return c
		}
		return ClassOther
	default:
		return ClassOther
	}
}

// errored reports whether a tool event recorded a failed outcome.
func errored(event trace.Event) bool {
	if event.Kind != "tool" {
		return false
	}
	var p trace.ToolPayload
	return trace.DecodePayload(event, &p) == nil && p.Status == "error"
}

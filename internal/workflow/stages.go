package workflow

import "skilltrace/internal/trace"

type Stage string

const (
	StageInvoke Stage = "invoke"
	StagePlan   Stage = "plan"
	StageAct    Stage = "act"
	StageVerify Stage = "verify"
	StageFinish Stage = "finish"
)

var StageOrder = []Stage{StageInvoke, StagePlan, StageAct, StageVerify, StageFinish}

func Classify(event trace.Event) Stage {
	switch event.Kind {
	case "skill":
		return StageInvoke
	case "session":
		return StagePlan
	case "tool", "file":
		return StageAct
	case "usage":
		return StageVerify
	case "completion":
		return StageFinish
	default:
		return StageAct
	}
}

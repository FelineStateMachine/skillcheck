package workflow

type Compatibility struct {
	Actors, SkillEvidence, ToolCalls, FileEvents, Usage bool
}

func CompatibleCapabilities(left, right Cohort) Compatibility {
	return Compatibility{
		Actors:        left.Capabilities.Actors && right.Capabilities.Actors,
		SkillEvidence: left.Capabilities.SkillEvidence && right.Capabilities.SkillEvidence,
		ToolCalls:     left.Capabilities.ToolCalls && right.Capabilities.ToolCalls,
		FileEvents:    left.Capabilities.FileEvents && right.Capabilities.FileEvents,
		Usage:         left.Capabilities.Usage && right.Capabilities.Usage,
	}
}

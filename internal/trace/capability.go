package trace

type CapabilityProfile struct {
	Actors        bool `json:"actors"`
	SkillEvidence bool `json:"skill_evidence"`
	ToolCalls     bool `json:"tool_calls"`
	FileEvents    bool `json:"file_events"`
	Usage         bool `json:"usage"`
}

func CodexCapabilities() CapabilityProfile {
	return CapabilityProfile{Actors: false, SkillEvidence: true, ToolCalls: true, FileEvents: true, Usage: true}
}

func ClaudeCapabilities() CapabilityProfile {
	return CapabilityProfile{Actors: true, SkillEvidence: true, ToolCalls: true, FileEvents: true, Usage: true}
}

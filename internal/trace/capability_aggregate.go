package trace

// AggregateCapabilities returns only capabilities present in every input.
func AggregateCapabilities(profiles ...CapabilityProfile) CapabilityProfile {
	if len(profiles) == 0 {
		return CapabilityProfile{}
	}
	r := profiles[0]
	for _, p := range profiles[1:] {
		r.Actors = r.Actors && p.Actors
		r.SkillEvidence = r.SkillEvidence && p.SkillEvidence
		r.ToolCalls = r.ToolCalls && p.ToolCalls
		r.FileEvents = r.FileEvents && p.FileEvents
		r.Usage = r.Usage && p.Usage
	}
	return r
}

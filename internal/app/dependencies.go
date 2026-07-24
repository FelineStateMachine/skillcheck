package app

type Invalidations struct {
	Episodes  bool `json:"episodes"`
	Cohorts   bool `json:"cohorts"`
	Workflows bool `json:"workflows"`
	Findings  bool `json:"findings"`
}

func policyInvalidations(corrections, findings, cohorts, rollups int) Invalidations {
	return Invalidations{Episodes: corrections > 0, Cohorts: corrections > 0 || cohorts > 0, Workflows: corrections > 0 || rollups > 0, Findings: corrections > 0 || findings > 0 || rollups > 0}
}

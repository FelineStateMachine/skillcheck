package policy

type DocumentV1 struct {
	Version       int            `yaml:"version" json:"version"`
	Corrections   []Correction   `yaml:"corrections,omitempty" json:"corrections,omitempty"`
	Findings      []FindingRule  `yaml:"findings,omitempty" json:"findings,omitempty"`
	Cohorts       []Cohort       `yaml:"cohorts,omitempty" json:"cohorts,omitempty"`
	ActionRollups []ActionRollup `yaml:"action_rollups,omitempty" json:"action_rollups,omitempty"`
}

type Match struct {
	Skill  string `yaml:"skill,omitempty" json:"skill,omitempty"`
	Tier   string `yaml:"tier,omitempty" json:"tier,omitempty"`
	Actor  string `yaml:"actor,omitempty" json:"actor,omitempty"`
	Source string `yaml:"source,omitempty" json:"source,omitempty"`
	Model  string `yaml:"model,omitempty" json:"model,omitempty"`
}

type Correction struct {
	ID      string `yaml:"id" json:"id"`
	Enabled *bool  `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Match   Match  `yaml:"match" json:"match"`
	Label   string `yaml:"label" json:"label"`
}

type FindingRule struct {
	ID           string   `yaml:"id" json:"id"`
	Enabled      *bool    `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Kind         string   `yaml:"kind" json:"kind"`
	Title        string   `yaml:"title" json:"title"`
	Detail       string   `yaml:"detail" json:"detail"`
	Match        Match    `yaml:"match,omitempty" json:"match,omitempty"`
	MinSamples   int      `yaml:"min_samples,omitempty" json:"min_samples,omitempty"`
	Capabilities []string `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
}

type Cohort struct {
	ID      string `yaml:"id" json:"id"`
	Enabled *bool  `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Match   Match  `yaml:"match" json:"match"`
}

type ActionRollup struct {
	ID      string   `yaml:"id" json:"id"`
	Enabled *bool    `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Stage   string   `yaml:"stage" json:"stage"`
	Actions []string `yaml:"actions" json:"actions"`
}

func enabled(v *bool) bool { return v == nil || *v }

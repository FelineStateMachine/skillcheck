package skills

type Exposure struct {
	Harness string `json:"harness"`
	Scope   string `json:"scope"`
	Root    string `json:"root"`
}
type Skill struct {
	Identity  Identity   `json:"identity"`
	Exposures []Exposure `json:"exposures"`
}
type Issue struct {
	Entry  string `json:"entry"`
	Reason string `json:"reason"`
}
type Result struct {
	Skills []Skill `json:"skills"`
	Issues []Issue `json:"issues,omitempty"`
}

package generated

type Dataset struct {
	Version     string     `json:"version"`
	Title       string     `json:"title"`
	GeneratedAt string     `json:"generatedAt"`
	SkillAlias  string     `json:"skillAlias"`
	Freshness   string     `json:"freshness"`
	Caveats     []string   `json:"caveats"`
	Comparison  Comparison `json:"comparison"`
}

type Node struct {
	ID     string `json:"id"`
	Stage  string `json:"stage"`
	Action string `json:"action"`
	Count  int    `json:"count"`
}
type Edge struct {
	From     string   `json:"from"`
	To       string   `json:"to"`
	Count    int      `json:"count"`
	Variants []string `json:"variants"`
}
type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}
type Cohort struct {
	Alias   string   `json:"alias"`
	Members int      `json:"members"`
	Models  []string `json:"models"`
	Graph   Graph    `json:"graph"`
}
type Difference struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Stage      string `json:"stage"`
	Action     string `json:"action"`
	Detail     string `json:"detail"`
	LeftCount  int    `json:"leftCount"`
	RightCount int    `json:"rightCount"`
	Supported  bool   `json:"supported"`
}
type Comparison struct {
	Left           Cohort       `json:"left"`
	Right          Cohort       `json:"right"`
	Differences    []Difference `json:"differences"`
	ClaimsWithheld bool         `json:"claimsWithheld"`
	Revision       string       `json:"revision"`
}

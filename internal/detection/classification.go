package detection

type Classification struct {
	Automated Tier   `json:"automated"`
	Final     Tier   `json:"final"`
	Revision  string `json:"revision"`
}

func Classify(c Candidate) Classification {
	return Classification{Automated: c.Tier, Final: c.Tier, Revision: "detection-v1"}
}

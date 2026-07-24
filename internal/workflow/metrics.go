package workflow

type Metrics struct {
	Samples   int `json:"samples"`
	Variants  int `json:"variants"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
	Unknown   int `json:"unknown"`
}

const MinimumComparativeSamples = 5

func (m Metrics) Limited() bool { return m.Samples < MinimumComparativeSamples }

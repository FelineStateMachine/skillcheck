package policy

type Layer struct {
	Name     string
	Document *Document
}

// Resolve overlays definitions by kind and stable ID. Later layers win while
// retaining the first-seen order, making project/local composition predictable.
func Resolve(layers ...Layer) DocumentV1 {
	out := DocumentV1{Version: 1}
	corrections, findings, cohorts, rollups := map[string]int{}, map[string]int{}, map[string]int{}, map[string]int{}
	for _, layer := range layers {
		if layer.Document == nil {
			continue
		}
		for _, v := range layer.Document.V1.Corrections {
			overlay(&out.Corrections, corrections, v.ID, v)
		}
		for _, v := range layer.Document.V1.Findings {
			overlay(&out.Findings, findings, v.ID, v)
		}
		for _, v := range layer.Document.V1.Cohorts {
			overlay(&out.Cohorts, cohorts, v.ID, v)
		}
		for _, v := range layer.Document.V1.ActionRollups {
			overlay(&out.ActionRollups, rollups, v.ID, v)
		}
	}
	return out
}

func overlay[T any](values *[]T, indexes map[string]int, id string, value T) {
	if i, ok := indexes[id]; ok {
		(*values)[i] = value
		return
	}
	indexes[id] = len(*values)
	*values = append(*values, value)
}

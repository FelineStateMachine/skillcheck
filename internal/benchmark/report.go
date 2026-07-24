package benchmark

type Slice struct {
	Harness, Model, Tier string
	Expected, Matched    int
	Precision, Recall    float64
}
type Report struct {
	Fixtures int     `json:"fixtures"`
	Slices   []Slice `json:"slices"`
}

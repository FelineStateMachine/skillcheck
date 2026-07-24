package app

type Progress struct {
	Stage     string `json:"stage"`
	Completed int    `json:"completed"`
	Total     int    `json:"total"`
}
type ProgressFunc func(Progress)
